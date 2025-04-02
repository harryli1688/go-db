//go:build sqlite
// +build sqlite

package db

import (
	_ "github.com/mattn/go-sqlite3"
)

func init() {
	EnableSQLite3 = true
}
