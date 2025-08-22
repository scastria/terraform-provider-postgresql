# Resource: postgresql_role_default_role
Represents the default role to set as active when a role connects
## Example usage
```hcl
resource "postgresql_role" "Role" {
  name = "MyRole"
}
resource "postgresql_role" "DefaultRole" {
  name = "MyDefault"
}
resource "postgresql_role_default_role" "example" {
  role = postgresql_role.Role.name
  default = postgresql_role.DefaultRole.name
}
```
## Argument Reference
* `role` - **(Required, ForceNew, String)** The name of the role.
* `default` - **(Required, String)** The name of the default role. Must be a role that is granted to the role.
## Attribute Reference
* `id` - **(String)** Same as `role`
## Import
Role default roles can be imported using a proper value of `id` as described above
