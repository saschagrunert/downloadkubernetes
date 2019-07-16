package sqlite

import (
	"database/sql"

	"github.com/pkg/errors"
	// Import the sqlite3 bindings
	"github.com/chuckha/downloadkubernetes/models"
	_ "github.com/mattn/go-sqlite3"
)

const (
	flavor = "sqlite3"
)

// Store holds the database connection and functions to interact with the saved data.
type Store struct {
	db         *sql.DB
	userIDstmt *sql.Stmt
}

// NewStore connects to the db and returns a store or an error if any
func NewStore(database string) (*Store, error) {
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dl := &models.Download{}
	if _, err := db.Exec(dl.CreateTableIfNotExistsQueries(flavor)); err != nil {
		return nil, errors.WithStack(err)
	}
	ui := &models.UserID{}
	if _, err := db.Exec(ui.CreateTableIfNotExistsQueries(flavor)); err != nil {
		return nil, errors.WithStack(err)
	}
	uidstmt, err := db.Prepare(ui.InsertIntoPreparedStatements(flavor))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Store{
		db:         db,
		userIDstmt: uidstmt,
	}, nil

}

// SaveDownload writes the download to disk
func (s *Store) SaveDownload(download *models.Download) error {
	return nil
}

// SaveUserID writes the UserID to disk
func (s *Store) SaveUserID(userID *models.UserID) error {
	r, err := s.userIDstmt.Exec(userID.ID, userID.CreateTime, userID.ExpireTime)
	if err != nil {
		return errors.WithStack(err)
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return errors.WithStack(err)
	}
	if affected != 1 {
		return errors.Errorf("more or less than 1 row affected: %d", affected)
	}
	return nil
}
