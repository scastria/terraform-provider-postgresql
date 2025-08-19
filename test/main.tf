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
# resource "postgresql_user_role" "UserRole" {
#   user = postgresql_user.User.name
#   role = postgresql_role.Role.name
# }
# resource "postgresql_user_default_role" "UserDefaultRole" {
#   user = postgresql_user.User.name
#   role = postgresql_role.Role.name
# }
# resource "postgresql_role_permission" "RolePermission" {
#   role = postgresql_role.Role.name
#   privilege = "CREATE"
# }
# resource "postgresql_role_permission" "RolePermission2" {
#   role = postgresql_role.Role.name
#   privilege = "GRANT OPTION"
# }
