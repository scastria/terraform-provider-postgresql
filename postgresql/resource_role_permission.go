package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	_ "github.com/lib/pq"
	"github.com/scastria/terraform-provider-postgresql/postgresql/client"
)

const (
	// Levels
	GLOBAL               = "global"
	DATABASE             = "database"
	TYPE                 = "type"
	TABLESPACE           = "tablespace"
	SCHEMA               = "schema"
	PARAMETER            = "parameter"
	LARGE_OBJECT         = "large object"
	LANGUAGE             = "language"
	FUNCTION             = "function"
	PROCEDURE            = "procedure"
	ROUTINE              = "routine"
	ALL_FUNCTIONS        = "all functions in schema"
	ALL_PROCEDURES       = "all procedures in schema"
	ALL_ROUTINES         = "all routines in schema"
	FOREIGN_SERVER       = "foreign server"
	FOREIGN_DATA_WRAPPER = "foreign data wrapper"
	DOMAIN               = "domain"
	SEQUENCE             = "sequence"
	ALL_SEQUENCES        = "all sequences in schema"
	TABLE                = "table"
	ALL_TABLES           = "all tables in schema"

	// Privileges
	SELECT         = "select"
	INSERT         = "insert"
	UPDATE         = "update"
	DELETE         = "delete"
	TRUNCATE       = "truncate"
	REFERENCES     = "references"
	TRIGGER        = "trigger"
	CREATE         = "create"
	CONNECT        = "connect"
	TEMPORARY      = "temporary"
	EXECUTE        = "execute"
	USAGE          = "usage"
	SET            = "set"
	ALTER_SYSTEM   = "alter system"
	ALL_PRIVILEGES = "all privileges"
	SUPERUSER      = "superuser"
	CREATE_DB      = "createdb"
	CREATE_ROLE    = "createrole"
	BYPASS_RLS     = "bypassrls"
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
			"database": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"privilege": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER, CREATE, CONNECT, TEMPORARY, EXECUTE, USAGE, SET, ALTER_SYSTEM, ALL_PRIVILEGES, SUPERUSER, CREATE_DB, CREATE_ROLE, BYPASS_RLS}, false),
			},
			"level": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"target"},
				Default:      GLOBAL,
				ValidateFunc: validation.StringInSlice([]string{GLOBAL, DATABASE, DOMAIN, FOREIGN_DATA_WRAPPER, FOREIGN_SERVER, LANGUAGE, LARGE_OBJECT, PARAMETER, SCHEMA, TABLESPACE, TYPE, FUNCTION, PROCEDURE, ROUTINE, ALL_FUNCTIONS, ALL_PROCEDURES, ALL_ROUTINES, SEQUENCE, ALL_SEQUENCES, TABLE, ALL_TABLES}, false),
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
	database := d.Get("database").(string)
	privilege := d.Get("privilege").(string)
	level := d.Get("level").(string)
	target := d.Get("target").(string)
	var query string
	var err error
	if level == GLOBAL {
		query, _, err = c.Exec(ctx, "", "alter role \"%s\" %s", role, privilege)
	} else {
		query, _, err = c.Exec(ctx, database, "grant %s on %s %s to \"%s\"", privilege, level, target, role)
	}
	if err != nil {
		d.SetId("")
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId(fmt.Sprintf("%s:%s:%s:%s:%s", role, database, privilege, level, target))
	return diags
}

func hasPrivilege(ctx context.Context, c *client.Client, role string, database string, privilege string, level string, target string) (bool, error) {
	if level == GLOBAL {
		var super, createdb, createrole, bypass bool
		query, row, err := c.QueryRow(ctx, "", "select rolsuper, rolcreatedb, rolcreaterole, rolbypassrls from pg_catalog.pg_roles where rolname = '%s'", role)
		if err != nil {
			return false, err
		}
		err = row.Scan(&super, &createdb, &createrole, &bypass)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == SUPERUSER) && (!super) {
			return false, nil
		}
		if (privilege == CREATE_DB) && (!createdb) {
			return false, nil
		}
		if (privilege == CREATE_ROLE) && (!createrole) {
			return false, nil
		}
		if (privilege == BYPASS_RLS) && (!bypass) {
			return false, nil
		}
	} else if level == DATABASE {
		var hasCreate, hasConnect, hasTemporary bool
		query, row, err := c.QueryRow(ctx, "", "select has_database_privilege('%s', '%s', '%s'), has_database_privilege('%s', '%s', '%s'), has_database_privilege('%s', '%s', '%s')", role, target, CREATE, role, target, CONNECT, role, target, TEMPORARY)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasCreate, &hasConnect, &hasTemporary)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == CREATE) && (!hasCreate) {
			return false, nil
		}
		if (privilege == CONNECT) && (!hasConnect) {
			return false, nil
		}
		if (privilege == TEMPORARY) && (!hasTemporary) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && ((!hasTemporary) || (!hasConnect) || (!hasCreate)) {
			return false, nil
		}
	} else if level == DOMAIN {
		var hasUsage bool
		// domain is a special form of type so use has_type_privilege
		query, row, err := c.QueryRow(ctx, database, "select has_type_privilege('%s', '%s', '%s')", role, target, USAGE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasUsage)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == USAGE) && (!hasUsage) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && (!hasUsage) {
			return false, nil
		}
	} else if level == FOREIGN_DATA_WRAPPER {
		var hasUsage bool
		query, row, err := c.QueryRow(ctx, database, "select has_foreign_data_wrapper_privilege('%s', '%s', '%s')", role, target, USAGE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasUsage)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == USAGE) && (!hasUsage) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && (!hasUsage) {
			return false, nil
		}
	} else if level == FOREIGN_SERVER {
		var hasUsage bool
		query, row, err := c.QueryRow(ctx, database, "select has_server_privilege('%s', '%s', '%s')", role, target, USAGE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasUsage)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == USAGE) && (!hasUsage) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && (!hasUsage) {
			return false, nil
		}
	} else if level == LANGUAGE {
		var hasUsage bool
		query, row, err := c.QueryRow(ctx, database, "select has_language_privilege('%s', '%s', '%s')", role, target, USAGE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasUsage)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == USAGE) && (!hasUsage) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && (!hasUsage) {
			return false, nil
		}
	} else if level == LARGE_OBJECT {
		var hasSelect, hasUpdate bool
		query, row, err := c.QueryRow(ctx, database, "select has_large_object_privilege('%s', '%s', '%s'), has_large_object_privilege('%s', '%s', '%s')", role, target, SELECT, role, target, UPDATE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasSelect, &hasUpdate)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == SELECT) && (!hasSelect) {
			return false, nil
		}
		if (privilege == UPDATE) && (!hasUpdate) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && ((!hasSelect) || (!hasUpdate)) {
			return false, nil
		}
	} else if level == PARAMETER {
		var hasSet, hasAlter bool
		query, row, err := c.QueryRow(ctx, database, "select has_parameter_privilege('%s', '%s', '%s'), has_parameter_privilege('%s', '%s', '%s')", role, target, SET, role, target, ALTER_SYSTEM)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasSet, &hasAlter)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == SET) && (!hasSet) {
			return false, nil
		}
		if (privilege == ALTER_SYSTEM) && (!hasAlter) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && ((!hasSet) || (!hasAlter)) {
			return false, nil
		}
	} else if level == SCHEMA {
		var hasCreate, hasUsage bool
		query, row, err := c.QueryRow(ctx, database, "select has_schema_privilege('%s', '%s', '%s'), has_schema_privilege('%s', '%s', '%s')", role, target, CREATE, role, target, USAGE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasCreate, &hasUsage)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == CREATE) && (!hasCreate) {
			return false, nil
		}
		if (privilege == USAGE) && (!hasUsage) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && ((!hasCreate) || (!hasUsage)) {
			return false, nil
		}
	} else if level == TABLESPACE {
		var hasCreate bool
		query, row, err := c.QueryRow(ctx, database, "select has_tablespace_privilege('%s', '%s', '%s')", role, target, CREATE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasCreate)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == CREATE) && (!hasCreate) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && (!hasCreate) {
			return false, nil
		}
	} else if level == TYPE {
		var hasUsage bool
		query, row, err := c.QueryRow(ctx, database, "select has_type_privilege('%s', '%s', '%s')", role, target, USAGE)
		if err != nil {
			return false, err
		}
		err = row.Scan(&hasUsage)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		if (privilege == USAGE) && (!hasUsage) {
			return false, nil
		}
		if (privilege == ALL_PRIVILEGES) && (!hasUsage) {
			return false, nil
		}
	}
	return true, nil
}

func resourceRolePermissionRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*client.Client)
	tokens := strings.Split(d.Id(), ":")
	role := tokens[0]
	database := tokens[1]
	privilege := tokens[2]
	level := tokens[3]
	target := tokens[4]
	hasPriv, err := hasPrivilege(ctx, c, role, database, privilege, level, target)
	if err != nil {
		d.SetId("")
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
	database := tokens[1]
	privilege := tokens[2]
	level := tokens[3]
	target := tokens[4]
	var query string
	var err error
	if level == GLOBAL {
		query, _, err = c.Exec(ctx, "", "alter role \"%s\" %s", role, fmt.Sprintf("no%s", privilege))
	} else {
		query, _, err = c.Exec(ctx, database, "revoke %s on %s %s from \"%s\"", privilege, level, target, role)
	}
	if err != nil {
		return diag.Errorf("Error executing query: %s, error: %v", query, err)
	}
	d.SetId("")
	return diags
}
