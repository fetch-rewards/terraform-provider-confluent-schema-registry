package schemaregistry

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"

	"github.com/ashleybill/srclient"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func resourceSchema() *schema.Resource {
	return &schema.Resource{
		CreateContext: schemaCreate,
		UpdateContext: schemaUpdate,
		ReadContext:   schemaRead,
		DeleteContext: schemaDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: customdiff.ComputedIf("version", func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) bool {

			var schemaHasChange bool
			oldState, newState := d.GetChange("schema")

			if schemaTypeStr, ok := d.Get("schema_type").(string); ok {
				if strings.ToLower(schemaTypeStr) == "json" {
					newJSON, _ := structure.NormalizeJsonString(newState)
					oldJSON, _ := structure.NormalizeJsonString(oldState)
					schemaHasChange = newJSON != oldJSON
				} else if strings.ToLower(schemaTypeStr) == "avro" {
					newJSON, _ := structure.NormalizeJsonString(newState)
					oldJSON, _ := structure.NormalizeJsonString(oldState)
					schemaHasChange = newJSON != oldJSON
				} else if strings.ToLower(schemaTypeStr) == "protobuf" {
					newProtoString, err := FormatProtobufString(newState.(string))
					if err != nil {
						// If theres an error diff should be true, indicating something is wrong?
						println("err")
					}
					oldProtoString, err := FormatProtobufString(oldState.(string))
					if err != nil {
						println("err")
					}

					schemaHasChange = oldProtoString != newProtoString
				}
			}
			log.Printf("[INFO] Schemas Equal %t", schemaHasChange)
			log.Printf("[INFO] Version Change %t", d.HasChange("version"))

			return schemaHasChange || d.HasChange("version")
		}),
		Schema: map[string]*schema.Schema{
			"subject": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The subject related to the schema",
				ForceNew:    true,
			},
			"schema": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The schema string",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					var schemaEquals bool

					if schemaTypeStr, ok := d.Get("schema_type").(string); ok {
						if strings.ToLower(schemaTypeStr) == "json" {
							newJSON, _ := structure.NormalizeJsonString(new)
							oldJSON, _ := structure.NormalizeJsonString(old)
							schemaEquals = newJSON == oldJSON
						} else if strings.ToLower(schemaTypeStr) == "avro" {
							newJSON, _ := structure.NormalizeJsonString(new)
							oldJSON, _ := structure.NormalizeJsonString(old)
							schemaEquals = newJSON == oldJSON
						} else if strings.ToLower(schemaTypeStr) == "protobuf" {
							newProtoString, err := FormatProtobufString(new)
							if err != nil {
								println("err")
							}

							oldProtoString, err := FormatProtobufString(old)
							if err != nil {
								println("err")
							}

							schemaEquals = newProtoString == oldProtoString
						}
					}
					return schemaEquals
				},
			},
			"schema_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The ID of the schema",
			},
			"version": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The schema string",
			},
			"reference": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The referenced schema list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The referenced schema name",
						},
						"subject": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The referenced schema subject",
						},
						"version": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The referenced schema version",
						},
					},
				},
			},
			"schema_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The schema type",
				Default:     "avro",
			},
		},
	}
}

func schemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	subject := d.Get("subject").(string)
	schemaString := d.Get("schema").(string)
	references := ToRegistryReferences(d.Get("reference").([]interface{}))
	schemaType := ToSchemaType(d.Get("schema_type"))

	client := meta.(*srclient.SchemaRegistryClient)

	schema, err := client.CreateSchema(subject, schemaString, schemaType, references...)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(formatSchemaVersionID(subject))
	d.Set("schema_id", schema.ID())
	d.Set("schema", schema.Schema())
	d.Set("version", schema.Version())

	if err = d.Set("reference", FromRegistryReferences(schema.References())); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func schemaUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	subject := d.Get("subject").(string)
	schemaString := d.Get("schema").(string)
	references := ToRegistryReferences(d.Get("reference").([]interface{}))
	schemaType := ToSchemaType(d.Get("schema_type"))
	currentSchemaId := d.Get("schema_id").(int)
	client := meta.(*srclient.SchemaRegistryClient)

	// This CreateSchema call does not fail if the schema already exists -- it just returns the schema.
	// This isn't ideal because if we update a schema with an OLD schema string, it will just return that old version
	// without updating the newest version to that version.
	// This results in a permanent diff in terraform -- because the latest schema is not matching what is in our new terraform.
	schema, err := client.CreateSchema(subject, schemaString, schemaType, references...)
	if err != nil {
		if strings.Contains(err.Error(), "409") {
			return diag.FromErr(fmt.Errorf("invalid 'schema': Incompatible. Please check the compatability level of your schema and compare it against the allowed actions found here: https://docs.confluent.io/cloud/current/sr/fundamentals/schema-evolution.html#compatibility-types."))
		}
		return diag.FromErr(err)
	}

	// If the schema returned from the above call is that of an EXISTING schema, we now do a soft delete on the old
	//schema and then recreate it, so that the "old" version is now the most updated version and our state matches
	//(soft delete just de-registers if from the subject it i think? but it still exists)
	if schema.ID() < currentSchemaId {
		err = client.DeleteSubjectByVersion(subject, schema.Version(), false)
		if err != nil {
			return diag.FromErr(err)
		}
		schema, err = client.CreateSchema(subject, schemaString, schemaType, references...)
		if err != nil {
			if strings.Contains(err.Error(), "409") {
				return diag.FromErr(fmt.Errorf("invalid 'schema': Incompatible. Please check the compatability level of your schema and compare it against the allowed actions found here: https://docs.confluent.io/cloud/current/sr/fundamentals/schema-evolution.html#compatibility-types."))
			}
			return diag.FromErr(err)
		}
	}
	d.Set("schema_id", schema.ID())
	d.Set("schema", schema.Schema())
	d.Set("version", schema.Version())

	if err = d.Set("reference", FromRegistryReferences(schema.References())); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func schemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*srclient.SchemaRegistryClient)
	subject := extractSchemaVersionID(d.Id())

	var err error
	// before, the provider tried to look up the schema by the schema string.
	// The issue was that when a terraform apply ran and failed, it was looking for a schema string that didn't exist (before the tf state gets updated even on a failure)
	// now, we do not use the tf state to refresh -- we get the latest schema from the registry
	latestSchema, err := client.GetLatestSchema(subject)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting last schema: %w", err))
	}

	// At this point, the schema read in matches the most recent version found in the kafka ui/registry
	d.Set("schema", latestSchema.Schema())
	d.Set("schema_id", latestSchema.ID())
	d.Set("subject", subject)
	d.Set("version", latestSchema.Version())

	if err = d.Set("reference", FromRegistryReferences(latestSchema.References())); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func schemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*srclient.SchemaRegistryClient)
	subject := extractSchemaVersionID(d.Id())

	err := client.DeleteSubject(subject, true)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func FromRegistryReferences(references []srclient.Reference) []interface{} {
	if len(references) == 0 {
		return make([]interface{}, 0)
	}

	refs := make([]interface{}, 0, len(references))
	for _, reference := range references {
		refs = append(refs, map[string]interface{}{
			"name":    reference.Name,
			"subject": reference.Subject,
			"version": reference.Version,
		})
	}

	return refs
}

func ToSchemaType(schemaType interface{}) srclient.SchemaType {
	returnType := srclient.Avro

	if schemaTypeStr, ok := schemaType.(string); ok {
		if strings.ToLower(schemaTypeStr) == "json" {
			returnType = srclient.Json
		} else if strings.ToLower(schemaTypeStr) == "protobuf" {
			returnType = srclient.Protobuf
		}
	}

	return returnType
}

func ToRegistryReferences(references []interface{}) []srclient.Reference {

	if len(references) == 0 {
		return make([]srclient.Reference, 0)
	}

	refs := make([]srclient.Reference, 0, len(references))
	for _, reference := range references {
		r := reference.(map[string]interface{})

		refs = append(refs, srclient.Reference{
			Name:    r["name"].(string),
			Subject: r["subject"].(string),
			Version: r["version"].(int),
		})
	}

	return refs
}
