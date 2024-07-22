package forum

import (
	"database/sql"
	"io/ioutil"
	"os"
)

func ExecuteSQLFile(db *sql.DB, filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	sqlBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	sqlCommands := string(sqlBytes)
	_, err = db.Exec(sqlCommands)
	if err != nil {
		return err
	}

	return nil
}