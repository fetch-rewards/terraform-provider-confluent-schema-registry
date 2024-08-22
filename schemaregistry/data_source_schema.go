package schemaregistry

import (
	"context"
	"fmt"

	"github.com/ashleybill/srclient"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceSchema() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSubjectRead,
		Schema: map[string]*schema.Schema{
			"subject": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The subject related to the schema",
			},
			"version": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The version of the schema",
			},
			"schema_id": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The schema ID",
			},
			"schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The schema string",
			},
			"references": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The referenced schema names list",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The referenced schema name",
						},
						"subject": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The subject related to the schema",
						},
						"version": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The version of the schema",
						},
					},
				},
			},
		},
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("schema", func(ctx context.Context, old, new, meta any) error {
				// If we are increasing "size" then the new value must be
				// a multiple of the old value.

				println(old.(string))
				println(new.(string))

				if new.(int) <= old.(int) {
					return nil
				}
				if (new.(int) % old.(int)) != 0 {
					return fmt.Errorf("new size value must be an integer multiple of old value %d", old.(int))
				}
				return nil
			}),
		),
	}
}

func dataSourceSubjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	subject := d.Get("subject").(string)
	version := d.Get("version").(int)

	client := m.(*srclient.SchemaRegistryClient)
	var schema *srclient.Schema
	var err error

	if version > 0 {
		schema, err = client.GetSchemaByVersion(subject, version)

	} else {
		schema, err = client.GetLatestSchema(subject)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("schema_id", schema.ID()); err != nil {
		return diag.FromErr(fmt.Errorf("error in dataSourceSubjectRead with setting schema_id: %w", err))
	}

	if err = d.Set("schema", schema.Schema()); err != nil {
		return diag.FromErr(fmt.Errorf("error in dataSourceSubjectRead with setting schema: %w", err))
	}

	if err = d.Set("version", schema.Version()); err != nil {
		return diag.FromErr(fmt.Errorf("error in dataSourceSubjectRead with setting version: %w", err))
	}

	if err = d.Set("references", FromRegistryReferences(schema.References())); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(formatSchemaVersionID(subject))

	return diags
}
