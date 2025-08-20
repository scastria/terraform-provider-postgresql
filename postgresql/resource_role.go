package postgresql

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"login": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"inherit": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	name := d.Get("name").(string)
	login := d.Get("login").(bool)
	inherit := d.Get("inherit").(bool)
	loginOption := ""
	if login {
		loginOption = "login"
	} else {
		loginOption = "nologin"
	}
	inheritOption := ""
	if inherit {
		inheritOption = "inherit"
	} else {
		inheritOption = "noinherit"
	}
	query, _, err := c.Exec(ctx, "create role \"%s\" with %s %s", name, loginOption, inheritOption)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId(name)
	return diags
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	name := d.Id()
	var login, inherit bool
	query, row := c.QueryRow(ctx, "select rolcanlogin, rolinherit from pg_catalog.pg_roles where rolname = '%s'", name)
	err := row.Scan(&login, &inherit)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.Set("name", name)
	d.Set("login", login)
	d.Set("inherit", inherit)
	return diags
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		query, _, err := c.Exec(ctx, "alter role \"%s\" rename to \"%s\"", oldName.(string), newName.(string))
		if err != nil {
			return diag.Errorf("Error executing query: %s, error: %v", query, err)
		}
		d.SetId(newName.(string))
	}
	name := d.Id()
	login := d.Get("login").(bool)
	inherit := d.Get("inherit").(bool)
	loginOption := ""
	if login {
		loginOption = "login"
	} else {
		loginOption = "nologin"
	}
	inheritOption := ""
	if inherit {
		inheritOption = "inherit"
	} else {
		inheritOption = "noinherit"
	}
	query, _, err := c.Exec(ctx, "alter role \"%s\" with %s %s", name, loginOption, inheritOption)
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	return diags
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	name := d.Id()
	query, _, err := c.Exec(ctx, "drop role \"%s\"", name)
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId("")
	return diags
}
