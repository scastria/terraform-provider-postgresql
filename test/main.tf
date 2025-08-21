terraform {
  required_providers {
    postgresql = {
      source = "github.com/scastria/postgresql"
    }
  }
}

provider "postgresql" {
}

# resource "postgresql_user" "User" {
#   name = "TestUser"
#   auth_plugin = "AWSAuthenticationPlugin"
#   auth_plugin_alias = "RDS"
#   email = "good@bad.com"
# }
# resource "postgresql_role" "Role" {
#   name = "TestRole"
# }
# resource "postgresql_role" "ParentRole" {
#   name = "ParentRole"
# }
# resource "postgresql_role_member" "RoleMember" {
#   role = postgresql_role.ParentRole.id
#   member = postgresql_role.Role.id
# }
# resource "postgresql_role_default_role" "DefaultRole" {
#   role = "MyRole"
#   default = "TestRole"
# }
# resource "postgresql_user_role" "UserRole" {
#   user = postgresql_user.User.name
#   role = postgresql_role.Role.name
# }
# resource "postgresql_user_default_role" "UserDefaultRole" {
#   user = postgresql_user.User.name
#   role = postgresql_role.Role.name
# }
resource "postgresql_role_permission" "RolePermission" {
  role = "test"
  database = "db_gdc_published"
  privilege = "select"
  level = "all sequences in schema"
  target = "sch_pro"
}
# resource "postgresql_role_permission" "RolePermission2" {
#   role = "test"
#   privilege = "createrole"
# }

# data "postgresql_databases" "DBs" {
# }
#
# output "test" {
#   value = data.postgresql_databases.DBs.names
# }
