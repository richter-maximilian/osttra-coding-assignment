package testhelpers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/RichterMaximilian/osttra-coding-assignment/migrate"
	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	dbContainer *PostgresContainer
	mngmtDBConn *pgxpool.Pool
	dbConnTmpl  string
	dbOnce      = &sync.Once{}
)

type CustomPool struct {
	*pgxpool.Pool
	db string
}

func (c *CustomPool) Close() error {
	c.Pool.Close()
	_, err := mngmtDBConn.Exec(context.Background(), fmt.Sprintf(`DELETE DATABASE "%s"`, c.db))
	return err
}

func GetDBPool(ctx context.Context) *CustomPool {
	dbOnce.Do(func() {
		var err error
		mngtDBConnStr := os.Getenv("DB_CONN")
		if mngtDBConnStr == "" {
			dbContainer, err = StartPostgres(ctx)
			if err != nil {
				panic(fmt.Errorf("starting postgres: %w", err))
			}
			mngtDBConnStr = dbContainer.ConnStr()
		}
		mngmtDBConn, err = pgxpool.Connect(ctx, mngtDBConnStr)
		if err != nil {
			if dbContainer != nil {
				dbContainer.Terminate(ctx)
			}
			panic(fmt.Errorf("connecting to DB: %w", err))
		}
		c := mngmtDBConn.Config().ConnConfig
		dbConnTmpl = fmt.Sprintf("postgres://%s:%s@%s:%d/%%s%s", c.User, c.Password, c.Host, c.Port, getConnStrQuery(c.ConnString()))
	})
	tmpDB := fmt.Sprintf("tpmdb-%s", uuid.NewString())
	_, err := mngmtDBConn.Exec(ctx, fmt.Sprintf(`CREATE DATABASE "%s"`, tmpDB))
	if err != nil {
		panic(fmt.Errorf("creating temporary DB: %w", err))
	}
	connStr := fmt.Sprintf(dbConnTmpl, tmpDB)
	pool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		panic(fmt.Errorf("connecting to db: %w", err))
	}
	return &CustomPool{Pool: pool, db: tmpDB}
}

func GetMigratedDBPool(ctx context.Context, migFile string) *CustomPool {
	pool := GetDBPool(ctx)
	err := migrate.Up(migFile, pool.Config().ConnString())
	if err != nil {
		panic(fmt.Errorf("migrating DB: %w", err))
	}
	return pool
}

func getConnStrQuery(connStr string) string {
	split := strings.Split(connStr, "?")
	if len(split) > 1 {
		return fmt.Sprintf("?%s", split[len(split)-1])
	}
	return ""
}

type PostgresContainer struct {
	testcontainers.Container
	port                                   nat.Port
	username, password, database, hostname string
}

func StartPostgres(ctx context.Context) (*PostgresContainer, error) {
	port, err := nat.NewPort("tcp", "5432")
	if err != nil {
		return nil, fmt.Errorf("unable to parse port: %w", err)
	}

	username, password, database := "test", "test", "test"
	hostname, exists := os.LookupEnv("TEST_DB_HOSTNAME")
	if !exists {
		hostname = "localhost"
	}
	req := testcontainers.ContainerRequest{
		Image:        "postgres:13-alpine",
		ExposedPorts: []string{fmt.Sprintf("%d/tcp", port.Int())},
		WaitingFor: wait.ForSQL(port, "pgx", func(port nat.Port) string {
			return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", username, password, hostname, port.Int(), database)
		}),
		Env: map[string]string{
			"POSTGRES_PASSWORD": password,
			"POSTGRES_USER":     username,
			"POSTGRES_DB":       database,
		},
	}
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to start postgres container: %w", err)
	}
	mappedPort, err := ctr.MappedPort(ctx, port)
	if err != nil {
		return nil, fmt.Errorf("finding mapped port: %w", err)
	}
	return &PostgresContainer{
		Container: ctr,
		port:      mappedPort,
		username:  username,
		password:  password,
		database:  database,
		hostname:  hostname,
	}, nil
}

func (p *PostgresContainer) ConnStr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", p.username, p.password, p.hostname, p.port.Int(), p.database)
}
