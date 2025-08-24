package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func dataSourceSchemas() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSchemasRead,
		Schema: map[string]*schema.Schema{
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"exclude": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceSchemasRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	database := d.Get("database").(string)
	exclude, ok := d.GetOk("exclude")
	excludeSet := schema.NewSet(schema.HashString, []interface{}{})
	if ok {
		excludeSet = exclude.(*schema.Set)
	}
	query, rows, err := c.Query(ctx, database, "select schema_name from information_schema.schemata order by schema_name")
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	names := []string{}
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
		if excludeSet.Contains(name) {
			continue
		}
		names = append(names, name)
	}
	d.Set("names", names)
	d.SetId(database)
	return diags
}
