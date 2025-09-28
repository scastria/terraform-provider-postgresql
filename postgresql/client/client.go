package client

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Client struct {
	host            string
	port            int
	defaultDatabase string
	username        string
	password        string
	conns           map[string]*sql.DB
	mu              sync.Mutex
}
type DatabaseNotExistError struct {
	Database string
}

func (e *DatabaseNotExistError) Error() string {
	return fmt.Sprintf("database %s does not exist", e.Database)
}

func NewClient(host string, port int, defaultDatabase string, username string, password string) (*Client, error) {
	c := &Client{
		host:            host,
		port:            port,
		defaultDatabase: defaultDatabase,
		username:        username,
		password:        password,
		conns:           make(map[string]*sql.DB),
	}
	return c, nil
}

func (c *Client) GetConn(database string) (*sql.DB, error) {
	if database == "" {
		database = c.defaultDatabase
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, exists := c.conns[database]
	if exists {
		return conn, nil
	}
	conn, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.username, c.password, c.host, c.port, database))
	if err != nil {
		return nil, err
	}
	// Verify database exists
	err = conn.Ping()
	if err != nil {
		return nil, &DatabaseNotExistError{Database: database}
	}
	c.conns[database] = conn
	return conn, nil
}

func (c *Client) QueryRow(ctx context.Context, database string, queryTemplate string, args ...any) (string, *sql.Row, error) {
	conn, err := c.GetConn(database)
	if err != nil {
		return "", nil, err
	}
	query := fmt.Sprintf(queryTemplate, args...)
	tflog.Info(ctx, "PostgreSQL SQL:", map[string]any{"SQL": query})
	return query, conn.QueryRow(query), nil
}

func (c *Client) Query(ctx context.Context, database string, queryTemplate string, args ...any) (string, *sql.Rows, error) {
	conn, err := c.GetConn(database)
	if err != nil {
		return "", nil, err
	}
	query := fmt.Sprintf(queryTemplate, args...)
	tflog.Info(ctx, "PostgreSQL SQL:", map[string]any{"SQL": query})
	rows, err := conn.Query(query)
	return query, rows, err
}

func (c *Client) Exec(ctx context.Context, database string, queryTemplate string, args ...any) (string, sql.Result, error) {
	conn, err := c.GetConn(database)
	if err != nil {
		return "", nil, err
	}
	query := fmt.Sprintf(queryTemplate, args...)
	tflog.Info(ctx, "PostgreSQL SQL:", map[string]any{"SQL": query})
	result, err := conn.Exec(query)
	return query, result, err
}
