package postgresql

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func resourceRoleMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleMemberCreate,
		ReadContext:   resourceRoleMemberRead,
		UpdateContext: resourceRoleMemberUpdate,
		DeleteContext: resourceRoleMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"member": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"inherit": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"set": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceRoleMemberCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	role := d.Get("role").(string)
	member := d.Get("member").(string)
	admin := d.Get("admin").(bool)
	inherit := d.Get("inherit").(bool)
	set := d.Get("set").(bool)
	query, _, err := c.Exec(ctx, "", "grant \"%s\" to \"%s\" with admin %t, inherit %t, set %t", role, member, admin, inherit, set)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId(fmt.Sprintf("%s:%s", role, member))
	return diags
}

func resourceRoleMemberRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	member := tokens[1]
	var admin, inherit, set bool
	query, row, err := c.QueryRow(ctx, "", "select m.admin_option, m.inherit_option, m.set_option from pg_catalog.pg_auth_members m join pg_catalog.pg_roles mr on m.member = mr.oid join pg_catalog.pg_roles r on m.roleid = r.oid where r.rolname = '%s' and mr.rolname = '%s'", role, member)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	err = row.Scan(&admin, &inherit, &set)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.Set("role", role)
	d.Set("member", member)
	d.Set("admin", admin)
	d.Set("inherit", inherit)
	d.Set("set", set)
	return diags
}

func resourceRoleMemberUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	member := tokens[1]
	admin := d.Get("admin").(bool)
	inherit := d.Get("inherit").(bool)
	set := d.Get("set").(bool)
	query, _, err := c.Exec(ctx, "", "grant \"%s\" to \"%s\" with admin %t, inherit %t, set %t", role, member, admin, inherit, set)
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	return diags
}

func resourceRoleMemberDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	member := tokens[1]
	query, _, err := c.Exec(ctx, "", "revoke \"%s\" from \"%s\"", role, member)
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId("")
	return diags
}
