# Data Source: postgresql_routines
Represents all routines in a schema
## Example usage
```hcl
data "postgresql_routines" "example" {
  database = "my_database"
  schema = "my_schema"
}
```
## Argument Reference
* `database` - **(Required, String)** The database to retrieve routines from.
* `schema` - **(Required, String)** The schema to retrieve routines from.
* `type` - **(Optional, String)** Type of routine to retrieve. Allowed values: `function`, `procedure`, `routine`. Default: `routine`.
## Attribute Reference
* `id` - **(String)** Same as`database`:`schema`
* `names` - **(List of String)** List of all routine names in `database` and `schema`.
