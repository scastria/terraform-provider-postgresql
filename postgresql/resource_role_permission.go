package postgresql

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

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
				ValidateFunc: validation.StringInSlice([]string{GLOBAL, DATABASE, DOMAIN, FOREIGN_DATA_WRAPPER, FOREIGN_SERVER, LANGUAGE, LARGE_OBJECT, PARAMETER, SCHEMA, TABLESPACE, TYPE, SEQUENCE, ALL_SEQUENCES, FUNCTION, ALL_FUNCTIONS, PROCEDURE, ALL_PROCEDURES, ROUTINE, ALL_ROUTINES, TABLE, ALL_TABLES}, false),
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

func fixCaseSensitiveIdentifier(identifier string) string {
	// Check if the string contains uppercase letters
	hasUppercase := strings.IndexFunc(identifier, unicode.IsUpper) != -1
	if !hasUppercase {
		return identifier
	}
	tokens := strings.Split(identifier, ".")
	for i, token := range tokens {
		tokens[i] = fmt.Sprintf("\"%s\"", token)
	}
	return strings.Join(tokens, ".")
}

func hasPrivilege(ctx context.Context, c *client.Client, role string, database string, privilege string, level string, target string) (bool, error) {
	targetCaseSensitive := fixCaseSensitiveIdentifier(target)
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
		query, row, err := c.QueryRow(ctx, "", "select has_database_privilege('%s', '%s', '%s'), has_database_privilege('%s', '%s', '%s'), has_database_privilege('%s', '%s', '%s')", role, targetCaseSensitive, CREATE, role, targetCaseSensitive, CONNECT, role, targetCaseSensitive, TEMPORARY)
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
		if (privilege == ALL_PRIVILEGES) && ((!hasCreate) || (!hasConnect) || (!hasTemporary)) {
			return false, nil
		}
	} else if level == DOMAIN {
		var hasUsage bool
		// domain is a special form of type so use has_type_privilege
		query, row, err := c.QueryRow(ctx, database, "select has_type_privilege('%s', '%s', '%s')", role, targetCaseSensitive, USAGE)
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
		query, row, err := c.QueryRow(ctx, database, "select has_foreign_data_wrapper_privilege('%s', '%s', '%s')", role, targetCaseSensitive, USAGE)
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
		query, row, err := c.QueryRow(ctx, database, "select has_server_privilege('%s', '%s', '%s')", role, targetCaseSensitive, USAGE)
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
		query, row, err := c.QueryRow(ctx, database, "select has_language_privilege('%s', '%s', '%s')", role, targetCaseSensitive, USAGE)
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
		query, row, err := c.QueryRow(ctx, database, "select has_large_object_privilege('%s', '%s', '%s'), has_large_object_privilege('%s', '%s', '%s')", role, targetCaseSensitive, SELECT, role, targetCaseSensitive, UPDATE)
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
		query, row, err := c.QueryRow(ctx, database, "select has_parameter_privilege('%s', '%s', '%s'), has_parameter_privilege('%s', '%s', '%s')", role, targetCaseSensitive, SET, role, targetCaseSensitive, ALTER_SYSTEM)
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
		query, row, err := c.QueryRow(ctx, database, "select has_schema_privilege('%s', '%s', '%s'), has_schema_privilege('%s', '%s', '%s')", role, targetCaseSensitive, CREATE, role, targetCaseSensitive, USAGE)
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
		query, row, err := c.QueryRow(ctx, database, "select has_tablespace_privilege('%s', '%s', '%s')", role, targetCaseSensitive, CREATE)
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
		query, row, err := c.QueryRow(ctx, database, "select has_type_privilege('%s', '%s', '%s')", role, targetCaseSensitive, USAGE)
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
	} else if level == SEQUENCE {
		hasPriv, err := hasSequencePrivilege(ctx, c, role, database, privilege, target)
		if err != nil {
			return false, err
		}
		if !hasPriv {
			return false, nil
		}
	} else if level == ALL_SEQUENCES {
		// Get all sequences
		query, rows, err := c.Query(ctx, database, "select sequence_name from information_schema.sequences where sequence_schema = '%s' order by sequence_name", target)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				return false, err
			}
			hasPriv, err := hasSequencePrivilege(ctx, c, role, database, privilege, fmt.Sprintf("%s.%s", target, name))
			if err != nil {
				return false, err
			}
			if !hasPriv {
				return false, nil
			}
		}
	} else if (level == FUNCTION) || (level == PROCEDURE) || (level == ROUTINE) {
		hasPriv, err := hasFunctionPrivilege(ctx, c, role, database, privilege, target)
		if err != nil {
			return false, err
		}
		if !hasPriv {
			return false, nil
		}
	} else if (level == ALL_FUNCTIONS) || (level == ALL_PROCEDURES) || (level == ALL_ROUTINES) {
		// Get all items
		var inFilter string
		if level == ALL_FUNCTIONS {
			inFilter = "'FUNCTION'"
		} else if level == ALL_PROCEDURES {
			inFilter = "'PROCEDURE'"
		} else {
			inFilter = "'FUNCTION', 'PROCEDURE'"
		}
		query, rows, err := c.Query(ctx, database, "select routine_name from information_schema.routines where routine_schema = '%s' and routine_type in (%s) order by routine_name", target, inFilter)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				return false, err
			}
			hasPriv, err := hasFunctionPrivilege(ctx, c, role, database, privilege, fmt.Sprintf("%s.%s", target, name))
			if err != nil {
				return false, err
			}
			if !hasPriv {
				return false, nil
			}
		}
	} else if level == TABLE {
		hasPriv, err := hasTablePrivilege(ctx, c, role, database, privilege, target)
		if err != nil {
			return false, err
		}
		if !hasPriv {
			return false, nil
		}
	} else if level == ALL_TABLES {
		// Get all tables and views
		query, rows, err := c.Query(ctx, database, "select table_name from information_schema.tables where table_schema = '%s' order by table_name", target)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
		}
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				return false, err
			}
			hasPriv, err := hasTablePrivilege(ctx, c, role, database, privilege, fmt.Sprintf("%s.%s", target, name))
			if err != nil {
				return false, err
			}
			if !hasPriv {
				return false, nil
			}
		}
	}
	return true, nil
}

func hasTablePrivilege(ctx context.Context, c *client.Client, role string, database string, privilege string, table string) (bool, error) {
	tableCaseSensitive := fixCaseSensitiveIdentifier(table)
	var hasSelect, hasInsert, hasUpdate, hasDelete, hasTruncate, hasReferences, hasTrigger bool
	query, row, err := c.QueryRow(ctx, database, "select has_table_privilege('%s', '%s', '%s'), has_table_privilege('%s', '%s', '%s'), has_table_privilege('%s', '%s', '%s'), has_table_privilege('%s', '%s', '%s'), has_table_privilege('%s', '%s', '%s'), has_table_privilege('%s', '%s', '%s'), has_table_privilege('%s', '%s', '%s')", role, tableCaseSensitive, SELECT, role, tableCaseSensitive, INSERT, role, tableCaseSensitive, UPDATE, role, tableCaseSensitive, DELETE, role, tableCaseSensitive, TRUNCATE, role, tableCaseSensitive, REFERENCES, role, tableCaseSensitive, TRIGGER)
	if err != nil {
		return false, err
	}
	err = row.Scan(&hasSelect, &hasInsert, &hasUpdate, &hasDelete, &hasTruncate, &hasReferences, &hasTrigger)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
	}
	if (privilege == SELECT) && (!hasSelect) {
		return false, nil
	}
	if (privilege == INSERT) && (!hasInsert) {
		return false, nil
	}
	if (privilege == UPDATE) && (!hasUpdate) {
		return false, nil
	}
	if (privilege == DELETE) && (!hasDelete) {
		return false, nil
	}
	if (privilege == TRUNCATE) && (!hasTruncate) {
		return false, nil
	}
	if (privilege == REFERENCES) && (!hasReferences) {
		return false, nil
	}
	if (privilege == TRIGGER) && (!hasTrigger) {
		return false, nil
	}
	if (privilege == ALL_PRIVILEGES) && ((!hasSelect) || (!hasInsert) || (!hasUpdate) || (!hasDelete) || (!hasTruncate) || (!hasReferences) || (!hasTrigger)) {
		return false, nil
	}
	return true, nil
}

func hasFunctionPrivilege(ctx context.Context, c *client.Client, role string, database string, privilege string, function string) (bool, error) {
	functionCaseSensitive := fixCaseSensitiveIdentifier(function)
	var hasExecute bool
	query, row, err := c.QueryRow(ctx, database, "select has_function_privilege('%s', '%s', '%s')", role, functionCaseSensitive, EXECUTE)
	if err != nil {
		return false, err
	}
	err = row.Scan(&hasExecute)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
	}
	if (privilege == EXECUTE) && (!hasExecute) {
		return false, nil
	}
	if (privilege == ALL_PRIVILEGES) && (!hasExecute) {
		return false, nil
	}
	return true, nil
}

func hasSequencePrivilege(ctx context.Context, c *client.Client, role string, database string, privilege string, sequence string) (bool, error) {
	sequenceCaseSensitive := fixCaseSensitiveIdentifier(sequence)
	var hasUsage, hasSelect, hasUpdate bool
	query, row, err := c.QueryRow(ctx, database, "select has_sequence_privilege('%s', '%s', '%s'), has_sequence_privilege('%s', '%s', '%s'), has_sequence_privilege('%s', '%s', '%s')", role, sequenceCaseSensitive, USAGE, role, sequenceCaseSensitive, SELECT, role, sequenceCaseSensitive, UPDATE)
	if err != nil {
		return false, err
	}
	err = row.Scan(&hasUsage, &hasSelect, &hasUpdate)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error executing query: %s, error: %v", query, err))
	}
	if (privilege == USAGE) && (!hasUsage) {
		return false, nil
	}
	if (privilege == SELECT) && (!hasSelect) {
		return false, nil
	}
	if (privilege == UPDATE) && (!hasUpdate) {
		return false, nil
	}
	if (privilege == ALL_PRIVILEGES) && ((!hasUsage) || (!hasSelect) || (!hasUpdate)) {
		return false, nil
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
