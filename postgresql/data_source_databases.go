package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func dataSourceDatabases() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDatabasesRead,
		Schema: map[string]*schema.Schema{
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

func dataSourceDatabasesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	query, rows, err := c.Query(ctx, "", "select datname from pg_catalog.pg_database where datistemplate = false order by datname")
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
		names = append(names, name)
	}
	d.Set("names", names)
	d.SetId("databases")
	return diags
}
