package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/chuckha/downloadkubernetes/events"
	"github.com/chuckha/downloadkubernetes/models"
	"github.com/pkg/errors"

	// Import the sqlite3 bindings
	_ "github.com/mattn/go-sqlite3"
)

const (
	flavor = "sqlite3"
)

// Store holds the database connection and functions to interact with the saved data.
type Store struct {
	db                *sql.DB
	insertQueries     map[string]*sql.Stmt
	userIDstmt        *sql.Stmt
	getRecentDL       *sql.Stmt
	saveDL            *sql.Stmt
	saveLinkCopyEvent *sql.Stmt
}

type creators interface {
	CreateTableIfNotExistsQueries(string) string
	InsertIntoPreparedStatements(string) string
	InsertQueryName() string
}

// NewStore connects to the db and returns a store or an error if any
func NewStore(database string) (*Store, error) {
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Create all tables if necessary for things that need tables
	creatingModels := []creators{
		&models.Download{},
		&models.UserID{},
		&events.LinkCopy{},
		&events.UserID{},
	}
	for _, ct := range creatingModels {
		if _, err := db.Exec(ct.CreateTableIfNotExistsQueries(flavor)); err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("%T", ct))
		}
	}

	insertQueries := map[string]*sql.Stmt{}
	// prepare insert statements
	for _, insert := range creatingModels {
		stmt, err := db.Prepare(insert.InsertIntoPreparedStatements(flavor))
		if err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("%+v", insert))
		}
		insertQueries[insert.InsertQueryName()] = stmt
	}

	// prepare fetch queries...might be harder than saves
	dl := &models.Download{}
	dlstmt, err := db.Prepare(dl.SelectRecentDownloads(flavor))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &Store{
		db:            db,
		insertQueries: insertQueries,
		getRecentDL:   dlstmt,
	}, nil
}

func (s *Store) save(queryName string, args ...interface{}) error {
	stmt, ok := s.insertQueries[queryName]
	if !ok {
		return errors.Errorf("unknown insert query %s", queryName)
	}
	r, err := stmt.Exec(args...)
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

// SaveDownload writes the download to disk
func (s *Store) SaveDownload(download *models.Download) error {
	return s.save(download.InsertQueryName(),
		download.User,
		download.Downloaded,
		download.OperatingSystem,
		download.Architecture,
		download.Version,
		download.Binary,
		download.URL,
	)
}

// SaveUserID writes the UserID to disk
func (s *Store) SaveUserID(userID *models.UserID) error {
	return s.save(userID.InsertQueryName(), userID.ID, userID.CreateTime, userID.ExpireTime)
}

func (s *Store) SaveCopyLinkEvent(evt *events.LinkCopy) error {
	return s.save(evt.InsertQueryName(), evt.UserID, evt.When, evt.URL)
}

func (s *Store) SaveUserIDEvent(evt *events.UserID) error {
	return s.save(evt.InsertQueryName(), evt.When, evt.UserID, evt.Action)
}

func (s *Store) GetRecentDownloads(userID *models.UserID, limit int) ([]*models.Download, error) {
	rows, err := s.getRecentDL.Query(userID.ID, limit)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	out := make([]*models.Download, limit)

	for rows.Next() {
		dl := models.Download{}
		if err := rows.Scan(&dl.OperatingSystem, &dl.Architecture, &dl.Version, &dl.Binary); err != nil {
			return nil, errors.WithStack(err)
		}
		out = append(out, &dl)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	return out, nil
}
