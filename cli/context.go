package cli

import (
	"context"
	"database/sql"

	"mini-kanban/db"
	"mini-kanban/db/dbgen"
)

// CmdContext holds common resources needed by CLI commands.
type CmdContext struct {
	DB          *sql.DB
	Queries     *dbgen.Queries
	ProjectID   int64
	ProjectName string
}

// Close closes the database connection.
func (c *CmdContext) Close() error {
	if c.DB != nil {
		return c.DB.Close()
	}
	return nil
}

// Context returns a background context.
func (c *CmdContext) Context() context.Context {
	return context.Background()
}

// NewCmdContext creates a new CmdContext with database and project initialized.
func NewCmdContext() (*CmdContext, error) {
	database, err := getDB()
	if err != nil {
		return nil, err
	}

	q := dbgen.New(database)
	projectID, projectName, err := getProjectID(q)
	if err != nil {
		database.Close()
		return nil, err
	}

	return &CmdContext{
		DB:          database,
		Queries:     q,
		ProjectID:   projectID,
		ProjectName: projectName,
	}, nil
}

// NewCmdContextWithDB creates a CmdContext with an existing database connection.
// Useful for testing or when DB is already open.
func NewCmdContextWithDB(database *sql.DB) (*CmdContext, error) {
	if err := db.RunMigrations(database); err != nil {
		return nil, err
	}

	q := dbgen.New(database)
	projectID, projectName, err := getProjectID(q)
	if err != nil {
		return nil, err
	}

	return &CmdContext{
		DB:          database,
		Queries:     q,
		ProjectID:   projectID,
		ProjectName: projectName,
	}, nil
}
