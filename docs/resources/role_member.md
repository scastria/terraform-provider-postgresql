# Resource: postgresql_role_member
Represents a role assigned to another role
## Example usage
```hcl
resource "postgresql_role" "ParentRole" {
  name = "MyParent"
}
resource "postgresql_role" "Role" {
  name = "MyRole"
}
resource "postgresql_role_member" "example" {
  role = postgresql_role.ParentRole.name
  member = postgresql_role.Role.name
}
```
## Argument Reference
* `role` - **(Required, ForceNew, String)** The name of the parent role, typically containing privileges.
* `member` - **(Required, ForceNew, String)** The name of the role.
* `admin` - **(Optional, Boolean)** Whether the member can in turn grant membership in the role to others, and revoke membership in the role as well. Default: `false`.
* `inherit` - **(Optional, Boolean)** Whether the member automatically has access to the privileges it is a member of. Default: `true`.
* `set` - **(Optional, Boolean)** Whether the member can change to the granted role. Default: `true`.
## Attribute Reference
* `id` - **(String)** Same as `role`:`member`
## Import
Role members can be imported using a proper value of `id` as described above
