package auth

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/akruszewski/awiki/settings"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

var UserNotFoundError = errors.New("User Not Found!")
var UserExistsError = errors.New("User with given username already exists.!")

// User structure used for communication with database. Strongly recommended to
// not serialize it for web api, it has Hash.
type DBUser struct {
	Username   string `json:"username" form:"username"`
	Repository string `json:"repository"`
	Hash       []byte `json:"Hash"`
}

type User struct {
	Username   string `json:"username" form:"username"`
	Password   string `json:"password" form:"password"`
	Repository string `json:"repository"`
}

type DBUsers []DBUser

type Authentifier interface {
	GenerateToken(userUUID string) (string, error)
	Authenticate(user, storedUser *User) bool
	Logout(tokenString string, token *jwt.Token) error
	Login(requestUser *User) (int, []byte)
}

// Fetches users from database.
func Users() (DBUsers, error) {
	data, err := ioutil.ReadFile(settings.UserDBPath)
	if err != nil {
		return DBUsers{}, errors.Wrap(err, "Can't read users db")
	}
	users := DBUsers{}
	err = json.Unmarshal(data, &users)
	if err != nil {
		return DBUsers{}, errors.Wrap(err, "Can't load users db")
	}
	return users, nil
}

// Fetches user from database.
func LoadUser(userName string) (DBUser, error) {
	users, err := Users()
	if err != nil {
		return DBUser{}, errors.Wrap(err, "Can't read user db")
	}
	for _, user := range users {
		if strings.Compare(user.Username, userName) == 0 {
			return user, nil
		}
	}
	return DBUser{}, UserNotFoundError
}

// Checks if user exists in db
func UserExists(userName string) bool {
	users, err := Users()
	if err != nil {
		return false
	}
	for _, user := range users {
		if strings.Compare(user.Username, userName) == 0 {
			return true
		}
	}
	return false
}

// Creates user with given password string and saves it to db.
func (u *DBUser) Save(password string) error {
	if UserExists(u.Username) {
		return UserExistsError
	}
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(strings.Join(
			[]string{settings.Secret, password, settings.Secret},
			"+",
		)),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return errors.Wrap(err, "Can't save user into db")
	}
	u.Hash = hash
	return nil
}
