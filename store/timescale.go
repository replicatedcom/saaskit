package store

import (
	"database/sql"

	"github.com/jackc/pgx/stdlib"
	"github.com/pkg/errors"
	"github.com/replicatedcom/saaskit/tracing/datadog"
)

var timescaleDB *sql.DB

type TimescaleOpts struct {
	URI string
}

func InitTimescale(opts TimescaleOpts) error {
	if opts.URI == "" {
		return errors.New("Timescale URI is not set")
	}

	RegisterDatadogDriver("pgx", &stdlib.Driver{}, "timescale")
	db, err := datadog.OpenSQL("pgx", opts.URI)
	if err != nil {
		return errors.Wrap(err, "open pgx")
	}

	timescaleDB = db
	return nil
}

func MustGetTimescaleSession() *sql.DB {
	if timescaleDB == nil {
		panic("Timescale is not initilized")
	}
	return timescaleDB
}
