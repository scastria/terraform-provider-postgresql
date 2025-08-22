# PostgreSQL Provider
The PostgreSQL provider is used to manage various administrative resources.  The provider
needs to be configured with the proper credentials before it can be used.

This provider does NOT cover 100% of the PostgreSQL features.  If there is something missing
that you would like to be added, please submit an Issue in corresponding GitHub repo.
## Example Usage
```hcl
terraform {
  required_providers {
    postgresql = {
      source  = "scastria/postgresql"
      version = "~> 0.1.0"
    }
  }
}

# Configure the PostgreSQL Provider
provider "postgresql" {
  host = "myserver.example.com"
  username = "XXXXX"
  password = "YYYYY"
}
```
## Argument Reference
* `host` - **(Required, String)** The hostname of the postgresql server. Can be specified via env variable `POSTGRESQL_HOST`.
* `port` - **(Optional, Integer)** The port of the postgresql server. Can be specified via env variable `POSTGRESQL_PORT`. Default: `5432`
* `default_database` - **(Optional, String)** The default database to connect to when no `database` is specified for database-specific resources. Can be specified via env variable `POSTGRESQL_DEFAULT_DATABASE`. Default: `postgres`.
* `username` - **(Required, String)** Username to connect to server as. Can be specified via env variable `POSTGRESQL_USERNAME`.
* `password` - **(Required, String)** Password to connect to server with. Can be specified via env variable `POSTGRESQL_PASSWORD`.
