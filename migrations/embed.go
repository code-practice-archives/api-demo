package migrations

import _ "embed"

//go:embed 001_create_users.sql
var CreateUsersMySQL string

//go:embed 001_create_users.sqlite.sql
var CreateUsersSQLite string
