package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
	pgacl "github.com/sean-/postgresql-acl"
)

const (
	// Levels
	TYPES     = "types"
	SCHEMAS   = "schemas"
	FUNCTIONS = "functions"
	ROUTINES  = "routines"
	SEQUENCES = "sequences"
	TABLES    = "tables"
)

func resourceRoleDefaultPermission() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceRoleDefaultPermissionCreate,
		ReadContext:   resourceRoleDefaultPermissionRead,
		DeleteContext: resourceRoleDefaultPermissionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"database": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"privilege": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{ALL_PRIVILEGES, CREATE, DELETE, EXECUTE, INSERT, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE, USAGE}, false),
			},
			"level": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{FUNCTIONS, ROUTINES, SCHEMAS, SEQUENCES, TABLES, TYPES}, false),
			},
			"creator": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"filter": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRoleDefaultPermissionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	role := d.Get("role").(string)
	database := d.Get("database").(string)
	privilege := d.Get("privilege").(string)
	level := d.Get("level").(string)
	creator := d.Get("creator").(string)
	creatorClause := ""
	if creator != "" {
		creatorClause = fmt.Sprintf("for role \"%s\"", creator)
	}
	filter := d.Get("filter").(string)
	filterClause := ""
	if filter != "" {
		filterClause = fmt.Sprintf("in schema %s", filter)
	}
	query, _, err := c.Exec(ctx, database, "alter default privileges %s %s grant %s on %s to \"%s\"", creatorClause, filterClause, privilege, level, role)
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId(fmt.Sprintf("%s:%s:%s:%s:%s:%s", role, database, privilege, level, creator, filter))
	return diags
}

func hasDefaultPrivilege(ctx context.Context, c *client.Client, role string, database string, privilege string, level string, creator string, filter string) (bool, error) {
	var creatorRole string
	if creator == "" {
		creatorRole = "current_role"
	} else {
		creatorRole = fmt.Sprintf("'%s'", creator)
	}
	filterClause := ""
	if filter != "" {
		filterClause = fmt.Sprintf("and defaclnamespace::regnamespace = '%s'::regnamespace", filter)
	}
	var objectType string
	switch level {
	case FUNCTIONS:
		objectType = "f"
	case ROUTINES:
		objectType = "f"
	case SCHEMAS:
		objectType = "n"
	case SEQUENCES:
		objectType = "S"
	case TABLES:
		objectType = "r"
	default:
		objectType = "T"
	}
	var privs pq.StringArray
	query, row, err := c.QueryRow(ctx, database, "select defaclacl from pg_catalog.pg_default_acl where defaclrole = %s::regrole %s and defaclobjtype = '%s'", creatorRole, filterClause, objectType)
	if err != nil {
		return false, err
	}
	err = row.Scan(&privs)
	if err != nil {
		return false, fmt.Errorf("Error executing query: %s, error: %w", query, err)
	}
	if privs == nil {
		return false, nil
	}
	for _, priv := range privs {
		acl, err := pgacl.Parse(priv)
		if err != nil {
			return false, fmt.Errorf("Error parsing ACL: %s, error: %w", priv, err)
		}
		if acl.Role != pq.QuoteIdentifier(role) {
			continue
		}
		if !acl.GetPrivilege(getDefaultPrivilegeSet(privilege, level)) {
			continue
		}
		return true, nil
	}
	return false, nil
}

func getDefaultPrivilegeSet(privilege string, level string) pgacl.Privileges {
	switch privilege {
	case CREATE:
		return pgacl.Create
	case DELETE:
		return pgacl.Delete
	case EXECUTE:
		return pgacl.Execute
	case INSERT:
		return pgacl.Insert
	case REFERENCES:
		return pgacl.References
	case SELECT:
		return pgacl.Select
	case TRIGGER:
		return pgacl.Trigger
	case TRUNCATE:
		return pgacl.Truncate
	case UPDATE:
		return pgacl.Update
	case USAGE:
		return pgacl.Usage
	case ALL_PRIVILEGES:
		switch level {
		case FUNCTIONS:
			return pgacl.Execute
		case ROUTINES:
			return pgacl.Execute
		case SCHEMAS:
			return pgacl.Create | pgacl.Usage
		case SEQUENCES:
			return pgacl.Usage | pgacl.Select | pgacl.Update
		case TABLES:
			return pgacl.Insert | pgacl.Select | pgacl.Update | pgacl.Delete | pgacl.Truncate | pgacl.References | pgacl.Trigger
		case TYPES:
			return pgacl.Usage
		}
	}
	return pgacl.NoPrivs
}

func resourceRoleDefaultPermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	database := tokens[1]
	privilege := tokens[2]
	level := tokens[3]
	creator := tokens[4]
	filter := tokens[5]
	hasPriv, err := hasDefaultPrivilege(ctx, c, role, database, privilege, level, creator, filter)
	if err != nil {
		d.SetId("")
		var dneErr *client.DatabaseNotExistError
		if errors.As(err, &dneErr) {
			// Database does not exist, so the permission cannot exist
			return diags
		}
		return diag.FromErr(err)
	}
	if !hasPriv {
		d.SetId("")
		return diags
	}
	d.Set("role", role)
	if database != "" {
		d.Set("database", database)
	}
	d.Set("privilege", privilege)
	d.Set("level", level)
	if creator != "" {
		d.Set("creator", creator)
	}
	if filter != "" {
		d.Set("filter", filter)
	}
	return diags
}

func resourceRoleDefaultPermissionDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	database := tokens[1]
	privilege := tokens[2]
	level := tokens[3]
	creator := tokens[4]
	creatorClause := ""
	if creator != "" {
		creatorClause = fmt.Sprintf("for role \"%s\"", creator)
	}
	filter := tokens[5]
	filterClause := ""
	if filter != "" {
		filterClause = fmt.Sprintf("in schema %s", filter)
	}
	query, _, err := c.Exec(ctx, database, "alter default privileges %s %s revoke %s on %s from \"%s\"", creatorClause, filterClause, privilege, level, role)
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId("")
	return diags
}
