package postgresql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func dataSourceSequences() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSequencesRead,
		Schema: map[string]*schema.Schema{
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schema": {
				Type:     schema.TypeString,
				Required: true,
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

func dataSourceSequencesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	database := d.Get("database").(string)
	schemaName := d.Get("schema").(string)
	query, rows, err := c.Query(ctx, database, "select sequencename from pg_catalog.pg_sequences where schemaname = '%s' order by sequencename", schemaName)
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
	d.SetId(fmt.Sprintf("%s:%s", database, schemaName))
	return diags
}
