package db

import (
	"xorm.io/xorm/migrate"
)

// Migrations for db migrate
var migrations = make([]*migrate.Migration, 0)

func AddMigrations(m ...*migrate.Migration) {
	migrations = append(migrations, m...)
}
