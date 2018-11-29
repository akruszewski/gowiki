package tiedot

import (
	"encoding/json"
	"fmt"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/akruszewski/awiki/settings"
	"golang.org/x/crypto/bcrypt"
)

func runQuery(query string, col *db.Col) (map[int]struct{}, error) {
	result := make(map[int]struct{})
	var jq interface{}
	if err := json.Unmarshal([]byte(query), &jq); err != nil {
		return nil, err
	}
	return result, db.EvalQuery(jq, col, &result)
}

// Connects to database, if db isn't exists create it. Returns connection
func Connect() (*db.DB, error) {
	wikiDB, err := db.OpenDB(settings.DBPath)
	if err != nil {
		return nil, err
	}
	return wikiDB, err
}

func InitUserCollection(userDB *db.DB) error {
	if err := userDB.Create("User"); err != nil {
		return err
	}
	col := userDB.Use("User")
	col.Index([]string{"username"})
	return nil
}

// Fetches user from database.
func User(userName string, userDB *db.DB) (*db.Col, error) {
	col := userDB.Use("Users")
	user, err := runQuery(fmt.Sprintf(`{"username": %s}`, userName), col)[0]
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Creates user and saves it to db.
func createUser(userName, password string, userDB *db.DB) (int, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return -1, err
	}

	var jsonDoc map[string]interface{}
	if err := json.Unmarshal(
		[]byte(fmt.Sprintf(
			`{"username": %s, "password": %s}`,
			userName,
			hash,
		)),
		&jsonDoc,
	); err != nil {
		return -1, err
	}

	col := userDB.Use("Users")
	col.Insert(jsonDoc)
	return
}
