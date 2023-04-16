package portfoliov1

import (
	"database/sql"
	"strings"

	"github.com/oxisto/money-gopher/persistence"
)

var _ persistence.StorageObject = &Security{}

func (*Security) InitTables(db *persistence.DB) (err error) {
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS securities (
name TEXT PRIMARY KEY,
display_name TEXT NOT NULL
);`)
	if err != nil {
		return err
	}

	return
}

func (*Security) PrepareReplace(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`REPLACE INTO securities (name, display_name) VALUES (?,?);`)
}

func (*Security) PrepareList(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name FROM securities`)
}

func (*Security) PrepareGet(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`SELECT name, display_name FROM securities WHERE name = ?`)
}

func (*Security) PrepareUpdate(db *persistence.DB, columns []string) (stmt *sql.Stmt, err error) {
	// We need to make sure to quote columns here because they are potentially evil user input
	var (
		query string
		set   []string
	)

	query = `UPDATE securities`
	set = make([]string, len(columns))
	for i, col := range columns {
		set[i] = "SET " + persistence.Quote(col) + " = ?"
	}

	query += " " + strings.Join(set, ", ") + " WHERE name = ?;"

	return db.Prepare(query)
}

func (*Security) PrepareDelete(db *persistence.DB) (stmt *sql.Stmt, err error) {
	return db.Prepare(`DELETE FROM securities WHERE name = ?`)
}

func (s *Security) ReplaceIntoArgs() []any {
	return []any{s.Name, s.DisplayName}
}

func (s *Security) UpdateArgs(columns []string) (args []any) {
	for _, col := range columns {
		switch col {
		case "name":
			args = append(args, s.Name)
		case "display_name":
			args = append(args, s.DisplayName)
		}
	}

	return args
}

func (*Security) Scan(sc persistence.Scanner) (obj persistence.StorageObject, err error) {
	var (
		s Security
	)

	err = sc.Scan(&s.Name, &s.DisplayName)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
