# Data Source: postgresql_sequences
Represents all sequences in a schema
## Example usage
```hcl
data "postgresql_sequences" "example" {
  database = "my_database"
  schema = "my_schema"
}
```
## Argument Reference
* `database` - **(Required, String)** The database to retrieve sequences from.
* `schema` - **(Required, String)** The schema to retrieve sequences from.
* `exclude` - **(Optional, List of String)** The sequence names to exclude from the result.
## Attribute Reference
* `id` - **(String)** Same as`database`:`schema`
* `names` - **(List of String)** List of all routine names in `database` and `schema`.
