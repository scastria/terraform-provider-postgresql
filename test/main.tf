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
  database = "prism"
  privilege = "select"
  level = "all tables in schema"
  target = "public"
}
# resource "postgresql_role_permission" "RolePermission2" {
#   role = "test"
#   privilege = "createrole"
# }
# resource "postgresql_role_default_permission" "RoleDefaultPermission" {
#   role = "aws-db-readers"
#   privilege = "all privileges"
#   level = "sequences"
#   creator = "aws-db-developers"
#   filter = "public"
# }
# data "postgresql_schemas" "SCHs" {
#   system = true
#   database = "db_research_unpublished"
# }
# output "test" {
#   value = data.postgresql_schemas.SCHs.names
# }

# resource "postgresql_role_restriction" "RoleRestriction" {
#   role = "test"
#   database = "db_gdc_published"
#   privilege = "select"
#   level = "sequence"
#   target = "sch_pro.tbl_db_scripts_run_log_auto_id_seq"
# }
