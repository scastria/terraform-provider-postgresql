# Data Source: postgresql_role
Represents a role
## Example usage
```hcl
data "postgresql_role" "example" {
  search_name = "admin"
}
```
## Argument Reference
* `search_name` - **(Optional, String)** The search string to apply to the name of the role. Uses contains.
* `name` - **(Optional, String)** The filter string to apply to the name of the role. Uses equality.
## Attribute Reference
* `id` - **(String)** Same as `name`
* `login` - **(Boolean)** Whether this role can perform a login.
* `inherit` - **(Boolean)** Whether this role inherits privileges from roles it is granted by default.
