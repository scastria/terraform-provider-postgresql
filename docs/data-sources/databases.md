# Data Source: postgresql_databases
Represents all databases on a server
## Example usage
```hcl
data "postgresql_databases" "example" {
}
```
## Argument Reference
* `exclude` - **(Optional, List of String)** The database names to exclude from the result.
## Attribute Reference
* `id` - **(String)** Fixed value of `databases`
* `names` - **(List of String)** List of all database names on the server
