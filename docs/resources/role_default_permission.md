# Resource: postgresql_role_default_permission
Represents a default permission of a role
## Example usage
```hcl
resource "postgresql_role" "Role" {
  name = "MyRole"
}
resource "postgresql_role" "CreateRole" {
  name = "MyCreateRole"
}
resource "postgresql_role_default_permission" "example" {
  role      = postgresql_role.Role.name
  privilege = "select"
  level     = "tables"
  creator   = postgresql_role.CreateRole.name
}
```
## Argument Reference
* `role` - **(Required, ForceNew, String)** The name of the role to give the default permission.
* `database` - **(Optional, ForceNew, String)** The database where the privilege is to be granted.
* `privilege` - **(Required, ForceNew, String)** The privilege to grant. Allowed values: `all privileges`, `create`, `delete`, `execute`, `insert`, `references`, `select`, `trigger`, `truncate`, `update`, `usage`
* `level` - **(Required, ForceNew, String)** At what level to grant the `privilege`. Allowed values: `functions`, `routines`, `schemas`, `sequences`, `tables`, `types`.
* `creator` - **(Optional, ForceNew, String)** The name of the role whose newly created objects should receive these default permissions. If omitted, the default permission applies to objects created by the username specified in the provider configuration.
* `filter` - **(Optional, ForceNew, String)** The name of the schema to limit which newly created objects should receive these default permissions.
## Attribute Reference
* `id` - **(String)** Same as `role`:`database`:`privilege`:`level`:`creator`:`filter`. Use empty string for parts of the id that do not apply.
## Import
Role default permissions can be imported using a proper value of `id` as described above
