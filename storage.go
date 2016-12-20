package tagstash

//go:generate sql/gen.sh

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	sqlcmd "github.com/aryszka/tagstash/sql"

	// package registers itself
	_ "github.com/lib/pq"

	// package registers itself
	_ "github.com/mattn/go-sqlite3"
)

const (
	sqlite   = "sqlite3"
	postgres = "postgres"

	// DefaultDriverName is used as the default sql driver (sqlite3).
	DefaultDriverName = sqlite

	// DefaultDataSourceName is used as the default data source (data.sqlite).
	DefaultDataSourceName = "data.sqlite"
)

type commands struct {
	createDB    string
	getEntries  string
	getTags     string
	insertEntry string
	deleteEntry string
	deleteTag   string
}

type storage struct {
	db       *sql.DB
	commands commands
}

func getCommands(driverName string) commands {
	c := commands{
		createDB:    sqlcmd.Cmd_create_db,
		getEntries:  sqlcmd.Cmd_get_entries,
		getTags:     sqlcmd.Cmd_get_tags,
		insertEntry: sqlcmd.Cmd_insert_entry,
		deleteEntry: sqlcmd.Cmd_delete_entry,
		deleteTag:   sqlcmd.Cmd_delete_tag,
	}

	if driverName == postgres {
		c.insertEntry = sqlcmd.Cmd_insert_entry_pq
	}

	return c
}

func newStorage(o StorageOptions) (*storage, error) {
	if o.DriverName == "" {
		o.DriverName = DefaultDriverName
	}

	if o.DataSourceName == "" {
		o.DataSourceName = DefaultDataSourceName
	}

	var initDB bool
	if o.DriverName == sqlite {
		if _, err := os.Stat(o.DataSourceName); os.IsNotExist(err) {
			initDB = true
		} else if err != nil {
			return nil, err
		}
	}

	db, err := sql.Open(o.DriverName, o.DataSourceName)
	if err != nil {
		return nil, err
	}

	c := getCommands(o.DriverName)

	if initDB {
		if _, err := db.Exec(c.createDB); err != nil {
			db.Close()
			return nil, err
		}
	}

	return &storage{
		db:       db,
		commands: c,
	}, nil
}

func (s *storage) Get(tags []string) ([]*Entry, error) {
	if len(tags) == 0 {
		return nil, nil
	}

	params := make([]string, len(tags))
	paramArgs := make([]interface{}, len(tags))
	for i := range params {
		params[i] = fmt.Sprintf("$%d", i+1)
		paramArgs[i] = tags[i]
	}

	paramString := strings.Join(params, ", ")
	r, err := s.db.Query(fmt.Sprintf(s.commands.getEntries, paramString), paramArgs...)
	if err != nil {
		return nil, err
	}

	var e []*Entry
	for r.Next() {
		var (
			tag, value string
			tagIndex   int
		)

		if err := r.Scan(&tag, &value, &tagIndex); err != nil {
			return nil, err
		}

		e = append(e, &Entry{
			Tag:      tag,
			Value:    value,
			TagIndex: tagIndex,
		})
	}

	return e, nil
}

func (s *storage) GetTags(value string) ([]string, error) {
	r, err := s.db.Query(s.commands.getTags, value)
	if err != nil {
		return nil, err
	}

	var tags []string
	for r.Next() {
		var tag string
		if err := r.Scan(&tag); err != nil {
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, err
}

func (s *storage) Set(e *Entry) error {
	_, err := s.db.Exec(s.commands.insertEntry, e.Tag, e.Value, e.TagIndex)
	return err
}

func (s *storage) Remove(e *Entry) error {
	_, err := s.db.Exec(s.commands.deleteEntry, e.Tag, e.Value)
	return err
}

func (s *storage) Delete(tag string) error {
	_, err := s.db.Exec(s.commands.deleteTag, tag)
	return err
}

func (s *storage) Close() {
	s.db.Close()
}
