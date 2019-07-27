package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chuckha/downloadkubernetes/events"
	"github.com/chuckha/downloadkubernetes/models"
	"github.com/pkg/errors"

	// Import the sqlite3 bindings
	_ "github.com/mattn/go-sqlite3"
)

const (
	flavor = "sqlite3"
)

type storeLogger interface {
	Debugf(string, ...interface{})
	Error(error)
}

// Store holds the database connection and functions to interact with the saved data.
type Store struct {
	log              storeLogger
	db               *sql.DB
	queries          map[string]*sql.Stmt
	fetchActiveUsers *sql.Stmt
}

type creators interface {
	CreateTableIfNotExistsQueries(string) string
	InsertIntoPreparedStatements(string) string
	InsertQueryName() string
}

// NewStore connects to the db and returns a store or an error if any
func NewStore(database string, l storeLogger) (*Store, error) {
	db, err := sql.Open("sqlite3", database)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// Create all tables if necessary for things that need tables
	creatingModels := []creators{
		&models.User{},
		&events.LinkCopy{},
		&events.UserID{},
	}
	for _, ct := range creatingModels {
		if _, err := db.Exec(ct.CreateTableIfNotExistsQueries(flavor)); err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("%T", ct))
		}
	}

	queries := map[string]*sql.Stmt{}
	// prepare insert statements
	for _, insert := range creatingModels {
		stmt, err := db.Prepare(insert.InsertIntoPreparedStatements(flavor))
		if err != nil {
			return nil, errors.Wrapf(err, fmt.Sprintf("%+v", insert))
		}
		queries[insert.InsertQueryName()] = stmt
	}

	id := &models.User{}
	expireUserIDStmt, err := db.Prepare(id.ExpireUserPreparedStatement(flavor))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	queries["expire-user"] = expireUserIDStmt

	fetchActiveStmt, err := db.Prepare(id.FetchActiveUsersClicksStatement(flavor))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Store{
		log:              l,
		db:               db,
		queries:          queries,
		fetchActiveUsers: fetchActiveStmt,
	}, nil
}

func (s *Store) exec(queryName string, args ...interface{}) error {
	stmt, ok := s.queries[queryName]
	if !ok {
		return errors.Errorf("unknown insert query %s", queryName)
	}
	s.log.Debugf("executing a query: %q", queryName)
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

// SaveUserID writes the UserID to disk
func (s *Store) SaveUser(user *models.User) error {
	return s.exec(user.InsertQueryName(), user.ID, user.CreateTime, user.ExpireTime)
}

// SaveCopyLinkEvent writes a copy link event to disk
func (s *Store) SaveCopyLinkEvent(evt *events.LinkCopy) error {
	return s.exec(evt.InsertQueryName(), evt.UserID, evt.When, evt.URL)
}

// SaveUserIDEvent saves a new user id event to disk
func (s *Store) SaveUserIDEvent(evt *events.UserID) error {
	return s.exec(evt.InsertQueryName(), evt.When, evt.User.ID, evt.Action)
}

func (s *Store) ExpireUser(id string) error {
	return s.exec("expire-user", fmt.Sprintf("%d", time.Now().Unix()), id)
}

func (s *Store) FetchClicksForUnexpiredUsers() ([]*events.LinkCopy, error) {
	rows, err := s.fetchActiveUsers.Query(fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	links := make([]*events.LinkCopy, 0)

	for rows.Next() {
		link := new(events.LinkCopy)
		if err := rows.Scan(&link.UserID, &link.URL, &link.When); err != nil {
			return nil, errors.WithStack(err)
		}
		links = append(links, link)
	}
	if err := rows.Close(); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return links, nil
}
