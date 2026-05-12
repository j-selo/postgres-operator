package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	databasev1 "github.com/j-selo/postgres-operator/api/v1"
)

type Client struct {
	conn *pgx.Conn
}

func New(ctx context.Context, connStr string) (*Client, error) {
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return &Client{conn: conn}, nil
}

func (c *Client) Close(ctx context.Context) {
	c.conn.Close(ctx)
}

func (c *Client) DatabaseExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := c.conn.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", name,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check database existence: %w", err)
	}
	return exists, nil
}

func (c *Client) Provision(ctx context.Context, spec databasev1.PostgresDatabaseSpec) error {
	// DDL identifiers cannot use $1 placeholders — sanitize them instead.
	dbIdent := pgx.Identifier{spec.Database}.Sanitize()
	userIdent := pgx.Identifier{spec.User}.Sanitize()

	if _, err := c.conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", dbIdent)); err != nil {
		return fmt.Errorf("create database: %w", err)
	}

	if _, err := c.conn.Exec(ctx, fmt.Sprintf("CREATE USER %s", userIdent)); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	// Password is a value, not an identifier — pass it as a parameter to avoid injection.
	if _, err := c.conn.Exec(ctx, fmt.Sprintf("ALTER USER %s PASSWORD $1", userIdent), spec.Password); err != nil {
		return fmt.Errorf("set password: %w", err)
	}

	if _, err := c.conn.Exec(ctx, fmt.Sprintf(
		"GRANT ALL PRIVILEGES ON DATABASE %s TO %s", dbIdent, userIdent,
	)); err != nil {
		return fmt.Errorf("grant privileges: %w", err)
	}

	return nil
}
