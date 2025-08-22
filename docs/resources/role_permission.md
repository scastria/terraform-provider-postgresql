# Resource: postgresql_role_permission
Represents a permission of a role
## Example usage
```hcl
resource "postgresql_role" "Role" {
  name = "MyRole"
}
resource "postgresql_role_permission" "example" {
  role    = postgresql_role.Role.id
  privilege  = "select"
  level = "table"
  target = "MyTable"
}
```
## Argument Reference
* `role` - **(Required, ForceNew, String)** The name of the role.
* `database` - **(Optional, ForceNew, String)** The database where the privilege is to be granted.  Required when granting privileges on database-specific objects.
* `privilege` - **(Required, ForceNew, String)** The privilege to grant. Allowed values: `all privileges`, `alter system`, `bypassrls`, `connect`, `create`, `createdb`, `createrole`, `delete`, `execute`, `insert`, `references`, `select`, `set`, `superuser`, `temporary`, `trigger`, `truncate`, `update`, `usage`
* `level` - **(Optional, ForceNew, String)** At what level to grant the `privilege`. Allowed values: `all functions in schema`, `all procedures in schema`, `all routines in schema`, `all sequences in schema`, `all tables in schema`, `database`, `domain`, `foreign data wrapper`, `foreign server`, `function`, `global`, `language`, `large object`, `parameter`, `procedure`, `routine`, `schema`, `sequence`, `table`, `tablespace`, `type`. Default: `global`.
* `target` - **(Optional, ForceNew, String)** The target of the `privilege`. Must be specified when `level` is NOT `global`. 
## Attribute Reference
* `id` - **(String)** Same as `role`:`database`:`privilege`:`level`:`target`. Use empty string for parts of the id that do not apply.
## Import
Role permissions can be imported using a proper value of `id` as described above
