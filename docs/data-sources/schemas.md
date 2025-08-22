# Data Source: postgresql_schemas
Represents all schemas in a database
## Example usage
```hcl
data "postgresql_schemas" "example" {
  database = "my_database"
}
```
## Argument Reference
* `database` - **(Required, String)** The database to retrieve schemas from.
## Attribute Reference
* `id` - **(String)** Same as`database`
* `names` - **(List of String)** List of all schema names in `database`
