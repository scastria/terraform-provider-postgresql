package postgresql

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

func resourceRolePermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRolePermissionCreate,
		ReadContext:   resourceRolePermissionRead,
		DeleteContext: resourceRolePermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"privilege": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"select", "insert", "update", "delete", "truncate", "references", "trigger", "create", "connect", "temporary", "execute", "usage", "set", "alter system", "all privileges", "superuser", "createdb", "createrole", "bypassrls"}, false),
			},
			"level": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"target"},
				Default:      "global",
				ValidateFunc: validation.StringInSlice([]string{"global", "type", "tablespace", "schema", "parameter", "large object", "language", "function", "procedure", "routine", "all functions in schema", "all procedures in schema", "all routines in schema", "foreign server", "foreign data wrapper", "domain", "database", "sequence", "all sequences in schema", "table", "all tables in schema"}, false),
			},
			"target": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"level"},
			},
		},
	}
}

func resourceRolePermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	role := d.Get("role").(string)
	privilege := d.Get("privilege").(string)
	level := d.Get("level").(string)
	targetRaw, ok := d.GetOk("target")
	var target string
	if ok {
		target = targetRaw.(string)
	} else {
		target = ""
	}
	var query string
	var err error
	if level == "global" {
		query, _, err = c.Exec(ctx, "alter role \"%s\" %s", role, privilege)
	} else {
		query, _, err = c.Exec(ctx, "grant %s on %s %s to \"%s\"", privilege, level, target, role)
	}
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId(fmt.Sprintf("%s:%s:%s:%s", role, privilege, level, target))
	return diags
}

func resourceRolePermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	privilege := tokens[1]
	level := tokens[2]
	target := tokens[3]
	var super, createdb, createrole, bypass bool
	query, row := c.QueryRow(ctx, "select rolsuper, rolcreatedb, rolcreaterole, rolbypassrls from pg_catalog.pg_roles where rolname = '%s'", role)
	err := row.Scan(&super, &createdb, &createrole, &bypass)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	if level == "global" {
		if (privilege == "superuser") && (!super) {
			d.SetId("")
			return diags
		}
		if (privilege == "createdb") && (!createdb) {
			d.SetId("")
			return diags
		}
		if (privilege == "createrole") && (!createrole) {
			d.SetId("")
			return diags
		}
		if (privilege == "bypassrls") && (!bypass) {
			d.SetId("")
			return diags
		}
	}
	d.Set("role", role)
	d.Set("privilege", privilege)
	d.Set("level", level)
	if target != "" {
		d.Set("target", target)
	}
	return diags
}

func resourceRolePermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	privilege := tokens[1]
	level := tokens[2]
	target := tokens[3]
	var query string
	var err error
	if level == "global" {
		query, _, err = c.Exec(ctx, "alter role \"%s\" %s", role, fmt.Sprintf("no%s", privilege))
	} else {
		query, _, err = c.Exec(ctx, "revoke %s on %s %s from \"%s\"", privilege, level, target, role)
	}
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId("")
	return diags
}
