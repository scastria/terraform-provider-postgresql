# Resource: postgresql_role
Represents a role
## Example usage
```hcl
resource "postgresql_role" "example" {
  name = "MyRole"
}
```
## Argument Reference
* `name` - **(Required, String)** The name of the role.
* `login` - **(Optional, Boolean)** Whether this role can perform a login. Default: `false`.
* `inherit` - **(Optional, Boolean)** Whether this role inherits privileges from roles it is granted by default. Default: `true`.
## Attribute Reference
* `id` - **(String)** Same as `name`.
## Import
Roles can be imported using a proper value of `id` as described above
