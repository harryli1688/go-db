package config

import "time"

type (
	// Config struct
	Config struct {
		Driver          string        `envconfig:"APP_DATABASE_DRIVER" default:"sqlite3"`
		Username        string        `envconfig:"APP_DATABASE_USERNAME" default:"root"`
		Password        string        `envconfig:"APP_DATABASE_PASSWORD" default:"root"`
		Name            string        `envconfig:"APP_DATABASE_NAME" default:"db"`
		Host            string        `envconfig:"APP_DATABASE_HOST" default:"localhost:3306"`
		SSLMode         string        `envconfig:"APP_DATABASE_SSLMODE" default:"disable"`
		Path            string        `envconfig:"APP_DATABASE_PATH" default:"data/db/test.db"`
		Schema          string        `envconfig:"APP_DATABASE_SCHEMA" default:"public"`
		Timeout         int           `envconfig:"APP_DATABASE_TIMEOUT" default:"500"`
		LogSQL          bool          `envconfig:"APP_DATABASE_LOG_SQL" default:"false"`
		MaxOpenConns    int           `envconfig:"APP_DATABASE_MAX_OPEN_CONNS"`
		MaxIdleConns    int           `envconfig:"APP_DATABASE_MAX_IDLE_CONNS" default:"2"`
		ConnMaxLifetime time.Duration `envconfig:"APP_DATABASE_CONN_MAX_LIFE_TIME"`
		UseSQLite3      bool          `envconfig:"APP_DATABASE_USE_SQLITE3"`
		UseMySQL        bool          `envconfig:"APP_DATABASE_USE_MYSQL"`
		UseMSSQL        bool          `envconfig:"APP_DATABASE_USE_MSSQL"`
		UsePostgreSQL   bool          `envconfig:"APP_DATABASE_USE_POSTGRESQL"`
		Charset         string        `envconfig:"APP_DATABASE_CHARSET" default:"utf8"`
		ConnectionURL   string        `envconfig:"APP_DATABASE_CONNECTION_URL"`
		SyncAndMigrate  bool          `envconfig:"APP_DATABASE_SYNC_AND_MIGRATE" default:"true"`
	}
)
