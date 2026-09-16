package db

import _ "embed"

// SchemaSQL contains the DDL for the ChoreSync database schema.
//
//go:embed schema.sql
var SchemaSQL string
