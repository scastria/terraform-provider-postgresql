package postgresql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func dataSourceRoutines() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoutinesRead,
		Schema: map[string]*schema.Schema{
			"database": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schema": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ROUTINE,
				ValidateFunc: validation.StringInSlice([]string{FUNCTION, PROCEDURE, ROUTINE}, false),
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

func dataSourceRoutinesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	database := d.Get("database").(string)
	schemaName := d.Get("schema").(string)
	routineType := d.Get("type").(string)
	typeClause := ""
	if routineType == FUNCTION {
		typeClause = "and prokind in ('f', 'a', 'w')"
	} else if routineType == PROCEDURE {
		typeClause = "and prokind = 'p'"
	}
	exclude, ok := d.GetOk("exclude")
	excludeSet := schema.NewSet(schema.HashString, []interface{}{})
	if ok {
		excludeSet = exclude.(*schema.Set)
	}
	query, rows, err := c.Query(ctx, database, "select proname from pg_catalog.pg_proc where pronamespace::regnamespace = '%s'::regnamespace %s order by proname", schemaName, typeClause)
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
	d.SetId(fmt.Sprintf("%s:%s", database, schemaName))
	return diags
}
