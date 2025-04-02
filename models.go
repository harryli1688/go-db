package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"gitea.mediatek.inc/go/db/config"
	otelxorm "gitea.mediatek.inc/go/otel-xorm"

	// Needed for the MySQL driver
	_ "github.com/go-sql-driver/mysql"

	// Needed for the Postgresql driver
	_ "github.com/lib/pq"

	"xorm.io/xorm"
	"xorm.io/xorm/migrate"
	"xorm.io/xorm/names"
)

var (
	tables []interface{}

	// HasEngine specifies if we have a xorm.Engine
	HasEngine bool

	// EnableSQLite3 for enable sqlite 3
	EnableSQLite3 bool

	// PageSize for item count per pag
	PageSize = 50
)

// Engine represents a xorm engine or session.
type Engine interface {
	Table(tableNameOrBean interface{}) *xorm.Session
	Count(...interface{}) (int64, error)
	Decr(column string, arg ...interface{}) *xorm.Session
	Delete(...interface{}) (int64, error)
	Exec(...interface{}) (sql.Result, error)
	Find(interface{}, ...interface{}) error
	Get(...interface{}) (bool, error)
	ID(interface{}) *xorm.Session
	In(string, ...interface{}) *xorm.Session
	Incr(column string, arg ...interface{}) *xorm.Session
	Insert(...interface{}) (int64, error)
	InsertOne(interface{}) (int64, error)
	Iterate(interface{}, xorm.IterFunc) error
	Join(joinOperator string, tablename interface{}, condition string, args ...interface{}) *xorm.Session
	SQL(interface{}, ...interface{}) *xorm.Session
	Where(interface{}, ...interface{}) *xorm.Session
	Asc(colNames ...string) *xorm.Session
	Limit(limit int, start ...int) *xorm.Session
	SumInt(bean interface{}, columnName string) (res int64, err error)
}

func AddTables(t ...any) {
	tables = append(tables, t...)
}

func GlobalInit() {
	gonicNames := []string{"SSL", "UID"}
	for _, name := range gonicNames {
		names.LintGonicMapper[name] = true
	}
}

func getPostgreSQLConnectionString(dbHost, dbUser, dbPasswd, dbName, dbParam, dbsslMode string) (connStr string) {
	host, port := parsePostgreSQLHostPort(dbHost)
	if host[0] == '/' { // looks like a unix socket
		if dbPasswd == "" {
			connStr = fmt.Sprintf("postgres://%s@:%s/%s%ssslmode=%s&host=%s",
				url.PathEscape(dbUser), port, dbName, dbParam, dbsslMode, host)
		} else {
			connStr = fmt.Sprintf("postgres://%s:%s@:%s/%s%ssslmode=%s&host=%s",
				url.PathEscape(dbUser), url.PathEscape(dbPasswd), port, dbName, dbParam, dbsslMode, host)
		}
	} else {
		if dbPasswd == "" {
			connStr = fmt.Sprintf("postgres://%s@%s:%s/%s%ssslmode=%s",
				url.PathEscape(dbUser), host, port, dbName, dbParam, dbsslMode)
		} else {
			connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s%ssslmode=%s",
				url.PathEscape(dbUser), url.PathEscape(dbPasswd), host, port, dbName, dbParam, dbsslMode)
		}
	}
	return
}

// ParseMSSQLHostPort splits the host into host and port
func ParseMSSQLHostPort(info string) (string, string) {
	host, port := "127.0.0.1", "1433"
	if strings.Contains(info, ":") {
		host = strings.Split(info, ":")[0]
		port = strings.Split(info, ":")[1]
	}
	if strings.Contains(info, ",") {
		host = strings.Split(info, ",")[0]
		port = strings.TrimSpace(strings.Split(info, ",")[1])
	}
	if len(info) > 0 {
		host = info
	}
	return host, port
}

func dbConnStr(config config.Config) (string, error) {
	connStr := ""
	Param := "?"
	if strings.Contains(config.Name, Param) {
		Param = "&"
	}
	switch config.Driver {
	case "mysql":
		connType := "tcp"
		if config.Host[0] == '/' { // looks like a unix socket
			connType = "unix"
		}
		tls := config.SSLMode
		if tls == "disable" { // allow (Postgres-inspired) default value to work in MySQL
			tls = "false"
		}
		connStr = fmt.Sprintf("%s:%s@%s(%s)/%s%scharset=%s&parseTime=true&tls=%s",
			config.Username,
			config.Password,
			connType,
			config.Host,
			config.Name, Param,
			config.Charset, tls)
	case "postgres":
		connStr = getPostgreSQLConnectionString(
			config.Host,
			config.Username,
			config.Password,
			config.Name, Param,
			config.SSLMode)
	case "mssql":
		host, port := ParseMSSQLHostPort(config.Host)
		connStr = fmt.Sprintf("server=%s; port=%s; database=%s; user id=%s; password=%s;",
			host, port,
			config.Name,
			config.Username,
			config.Password)
	case "sqlite3":
		if !EnableSQLite3 {
			return "", errors.New("this binary version does not build support for SQLite3")
		}
		if err := os.MkdirAll(path.Dir(config.Path), os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to create directories: %v", err)
		}
		connStr = fmt.Sprintf("file:%s?cache=shared&mode=rwc&_busy_timeout=%d&_txlock=immediate",
			config.Path,
			config.Timeout)
	default:
		return "", fmt.Errorf("unknown database type: %s", config.Driver)
	}

	return connStr, nil
}

func getEngine(config config.Config) (*xorm.Engine, error) {
	var connStr string
	var err error

	connStr = config.ConnectionURL
	if connStr == "" {
		connStr, err = dbConnStr(config)
		if err != nil {
			return nil, err
		}
	}

	engine, err := xorm.NewEngine(config.Driver, connStr)
	if err != nil {
		return nil, err
	}

	if config.Driver == "mysql" {
		engine.Dialect().SetParams(map[string]string{"rowFormat": "DYNAMIC"})
	}
	engine.SetSchema(config.Schema)
	return engine, nil
}

// parsePostgreSQLHostPort parses given input in various forms defined in
// https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING
// and returns proper host and port number.
func parsePostgreSQLHostPort(info string) (string, string) {
	host, port := "127.0.0.1", "5432"
	if strings.Contains(info, ":") && !strings.HasSuffix(info, "]") {
		idx := strings.LastIndex(info, ":")
		host = info[:idx]
		port = info[idx+1:]
	} else if len(info) > 0 {
		host = info
	}
	return host, port
}

// setEngine sets the xorm.Engine
func setEngine(cfg config.Config) (*xorm.Engine, error) {
	x, err := getEngine(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	x.SetMapper(names.GonicMapper{})
	// WARNING: for serv command, MUST remove the output to os.stdout,
	// so use log file to instead print to stdout.
	// x.SetLogger(log.XORMLogger)
	x.ShowSQL(cfg.LogSQL)

	x.SetMaxOpenConns(cfg.MaxOpenConns)
	x.SetMaxIdleConns(cfg.MaxIdleConns)
	x.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	x.AddHook(otelxorm.NewTracingHook())

	return x, nil
}

// NewEngine initializes a new xorm.Engine
func NewEngine(cfg config.Config) (*xorm.Engine, error) {
	GlobalInit()
	x, err := setEngine(cfg)
	if err != nil {
		return nil, err
	}

	if err = x.Ping(); err != nil {
		return nil, err
	}

	if cfg.SyncAndMigrate {
		if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
			return nil, fmt.Errorf("sync database struct error: %v", err)
		}

		m := migrate.New(x, migrate.DefaultOptions, migrations)
		if err = m.Migrate(); err != nil {
			return nil, fmt.Errorf("migrate: %v", err)
		}
	}

	return x, nil
}

func Create(ctx context.Context, x *xorm.Engine, i interface{}) (err error) {
	sess := x.NewSession()
	sess.Context(ctx)
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}
	if _, err = sess.Insert(i); err != nil {
		return err
	}
	return sess.Commit()
}

func Delete(ctx context.Context, x *xorm.Engine, i interface{}) (err error) {
	sess := x.NewSession()
	sess.Context(ctx)
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	if cnt, err := sess.Delete(i); err != nil {
		return err
	} else if cnt != 1 {
		return errors.New("id or uuid not exist")
	}
	return sess.Commit()
}

// Statistic contains the database statistics
type Statistic struct {
	Counter struct {
		User int64
	}
}
