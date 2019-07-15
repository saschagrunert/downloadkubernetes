package sqlite3

import (
	"github.com/chuckha/downloadkubernetes/models"
)

type Store struct {

}

func (s *Store) 	SaveDownload(*models.Download) error {
	return nil
}