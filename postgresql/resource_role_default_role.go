package postgresql

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func resourceRoleDefaultRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleDefaultRoleCreate,
		ReadContext:   resourceRoleDefaultRoleRead,
		UpdateContext: resourceRoleDefaultRoleUpdate,
		DeleteContext: resourceRoleDefaultRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"default": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceRoleDefaultRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	role := d.Get("role").(string)
	default_role := d.Get("default").(string)
	query, _, err := c.Exec(ctx, "", "alter role \"%s\" set role = '%s'", role, default_role)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId(role)
	return diags
}

func resourceRoleDefaultRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	role := d.Id()
	var rolconfig pq.StringArray
	query, row, err := c.QueryRow(ctx, "", "select rolconfig from pg_catalog.pg_roles where rolname = '%s'", role)
	if err != nil {
		d.SetId("")
		return diag.FromErr(err)
	}
	err = row.Scan(&rolconfig)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	if rolconfig == nil {
		d.SetId("")
		return diags
	}
	// Look for the default role setting
	default_role := ""
	for _, config := range rolconfig {
		if strings.HasPrefix(config, "role=") {
			default_role = strings.TrimPrefix(config, "role=")
			break
		}
	}
	if default_role == "" {
		d.SetId("")
		return diags
	}
	d.Set("role", role)
	d.Set("default", default_role)
	return diags
}

func resourceRoleDefaultRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	role := d.Id()
	default_role := d.Get("default").(string)
	query, _, err := c.Exec(ctx, "", "alter role \"%s\" set role = '%s'", role, default_role)
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	return diags
}

func resourceRoleDefaultRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	role := d.Id()
	query, _, err := c.Exec(ctx, "", "alter role \"%s\" set role = default", role)
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId("")
	return diags
}
