# Data Source: postgresql_views
Represents all views in a schema
## Example usage
```hcl
data "postgresql_views" "example" {
  database = "my_database"
  schema = "my_schema"
}
```
## Argument Reference
* `database` - **(Required, String)** The database to retrieve views from.
* `schema` - **(Required, String)** The schema to retrieve views from.
## Attribute Reference
* `id` - **(String)** Same as`database`:`schema`
* `names` - **(List of String)** List of all routine names in `database` and `schema`.
