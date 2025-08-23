# Data Source: postgresql_tables
Represents all tables in a schema
## Example usage
```hcl
data "postgresql_tables" "example" {
  database = "my_database"
  schema = "my_schema"
}
```
## Argument Reference
* `database` - **(Required, String)** The database to retrieve tables from.
* `schema` - **(Required, String)** The schema to retrieve tables from.
## Attribute Reference
* `id` - **(String)** Same as`database`:`schema`
* `names` - **(List of String)** List of all routine names in `database` and `schema`.
