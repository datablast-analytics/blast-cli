package snowflake

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/snowflakedb/gosnowflake"
	"go.uber.org/zap"
)

const (
	invalidQueryErrorPrefix = "001003 (42000): SQL compilation error"
)

type DB struct {
	conn   *sqlx.DB
	logger *zap.SugaredLogger
}

func NewDB(c *gosnowflake.Config, logger *zap.SugaredLogger) (*DB, error) {
	dsn, err := gosnowflake.DSN(c)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DSN")
	}

	gosnowflake.GetLogger().SetOutput(ioutil.Discard)

	db, err := sqlx.Connect("snowflake", dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to snowflake")
	}

	return &DB{conn: db, logger: logger}, nil
}

func (db DB) IsValid(ctx context.Context, query string) (bool, error) {
	rows, err := db.conn.QueryContext(ctx, fmt.Sprintf("EXPLAIN %s", query))
	if err != nil {
		errorMessage := err.Error()
		if strings.HasPrefix(errorMessage, invalidQueryErrorPrefix) {
			errorSegments := strings.Split(errorMessage, "\n")
			if len(errorSegments) > 1 {
				err = errors.New(errorSegments[1])
			}
		}
	}

	if rows != nil {
		defer rows.Close()
	}

	return err == nil, err
}
