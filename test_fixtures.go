package db

import (
	"github.com/harryli1688/go-db/config"
	"github.com/harryli1688/go-otel-xorm"

	"github.com/go-testfixtures/testfixtures/v3"
	"xorm.io/xorm"
)

var fixtures *testfixtures.Loader

// LoadFixtures load fixtures for a test database
func LoadFixtures() error {
	return fixtures.Load()
}

// PrepareTestDatabase load test fixtures into test database
func PrepareTestDatabase() error {
	return LoadFixtures()
}

func CreateTestEngine(cfg config.Config, fixturesDir string) (*xorm.Engine, error) {
	var err error

	x, err := NewEngine(cfg)
	if err != nil {
		return nil, err
	}

	fixtures, err = testfixtures.New(
		testfixtures.Database(x.DB().DB),
		testfixtures.Dialect(cfg.Driver),
		testfixtures.Directory(fixturesDir),
	)

	return x, err
}
