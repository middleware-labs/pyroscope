package chstore

import (
	"context"
	"database/sql"
	"fmt"

	chgorm "gorm.io/driver/clickhouse"

	"gorm.io/gorm"

	// "github.com/pyroscope-io/pyroscope/pkg/chstore/migrations"
	"github.com/pyroscope-io/pyroscope/pkg/config"
	// "gorm.io/gorm"
)

type CHStore struct {
	config *config.Server

	db  *sql.DB
	orm *gorm.DB
}

func Open(c *config.Server) (*CHStore, error) {
	s := CHStore{config: c}

	dsn := "tcp://clickhouse:clickhouse@localhost:9000/default?read_timeout=10s"
	// dsn := "tcp://localhost:9000/default?username=clickhouse&password=clickhouse&read_timeout=10s"
	var err error
	s.orm, err = gorm.Open(chgorm.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	s.db, err = s.orm.DB()

	if err = s.Ping(context.Background()); err != nil {
		fmt.Println("connection error", err)
	} else {
		fmt.Println("connection success")
	}

	return &s, nil
}

func (s *CHStore) DB() *gorm.DB { return s.orm }

func (s *CHStore) Close() error { return s.db.Close() }

func (s *CHStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}
