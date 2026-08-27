package postgresql

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func dataSourceRole() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoleRead,
		Schema: map[string]*schema.Schema{
			"search_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"name"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"search_name"},
			},
			"login": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"inherit": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	var whereClause string
	searchName, ok := d.GetOk("search_name")
	if ok {
		whereClause = fmt.Sprintf("rolname like '%%%s%%'", searchName.(string))
	}
	name, ok := d.GetOk("name")
	if ok {
		whereClause = fmt.Sprintf("rolname = '%s'", name.(string))
	}
	query, rows, err := c.Query(ctx, "", "select rolname, rolcanlogin, rolinherit from pg_catalog.pg_roles where %s", whereClause)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	foundRole := false
	var login, inherit bool
	var foundName string
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&foundName, &login, &inherit)
		if err != nil {
			d.SetId("")
			return diag.FromErr(err)
		}
		if foundRole {
			d.SetId("")
			return diag.FromErr(fmt.Errorf("Filter criteria does not result in a single role"))
		}
		foundRole = true
	}
	if !foundRole {
		d.SetId("")
		return diag.FromErr(fmt.Errorf("No role exists with that filter criteria"))
	}
	d.Set("name", foundName)
	d.Set("login", login)
	d.Set("inherit", inherit)
	d.SetId(foundName)
	return diags
}
