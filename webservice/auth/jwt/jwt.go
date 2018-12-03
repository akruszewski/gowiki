package jwt

import (
	"bufio"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	settings "github.com/akruszewski/awiki/settings"
	auth "github.com/akruszewski/awiki/webservice/auth"
	"github.com/akruszewski/awiki/webservice/auth/redis"
	jwt "github.com/dgrijalva/jwt-go"
	jwtRequst "github.com/dgrijalva/jwt-go/request"
	"golang.org/x/crypto/bcrypt"
)

type TokenAuthentication struct {
	Token string `json:"token" form:"token"`
}

const (
	tokenDuration = 72
	expireOffset  = 3600
)

type JWTAuthenticationBackend struct {
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

var authBackendInstance *JWTAuthenticationBackend = nil

func InitJWTAuthenticationBackend() *JWTAuthenticationBackend {
	if authBackendInstance == nil {
		authBackendInstance = &JWTAuthenticationBackend{
			privateKey: getPrivateKey(),
			PublicKey:  getPublicKey(),
		}
	}
	return authBackendInstance
}

func (backend *JWTAuthenticationBackend) GenerateToken(userUUID string) (string, error) {
	token := jwt.New(jwt.SigningMethodRS512)
	token.Claims = jwt.MapClaims{
		"exp": time.Now().Add(
			time.Hour * time.Duration(settings.JWTExpirationDelta),
		).Unix(),
		"iat": time.Now().Unix(),
		"sub": userUUID,
	}
	tokenString, err := token.SignedString(backend.privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (backend *JWTAuthenticationBackend) Authenticate(user *auth.User, storedUser *auth.DBUser) bool {
	return user.Username == storedUser.Username && bcrypt.CompareHashAndPassword(
		[]byte(storedUser.Hash),
		[]byte(user.Password),
	) == nil
}

func (backend *JWTAuthenticationBackend) getTokenRemainingValidity(timestamp interface{}) time.Duration {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			duration, _ := time.ParseDuration(fmt.Sprintf("%vs", remainer.Seconds()+expireOffset))
			return duration
		}
	}
	return expireOffset
}

func (backend *JWTAuthenticationBackend) Logout(tokenString string, token *jwt.Token) error {
	redisConn := redis.Connect()
	return redisConn.Set(
		tokenString,
		tokenString,
		backend.getTokenRemainingValidity(token.Claims.(jwt.MapClaims)["exp"]),
	).Err()
}

func (backend *JWTAuthenticationBackend) IsInBlockList(token string) bool {
	redisConn := redis.Connect()
	redisToken, err := redisConn.Get(token).Result()

	if err != nil || len(redisToken) < 1 {
		return false
	}

	return true
}

//Logins requestUser by checking it credentials compared to dbUser
func Login(requestUser *auth.User) (int, []byte) {
	authBackend := InitJWTAuthenticationBackend()

	dbUser, err := auth.LoadUser(requestUser.Username)
	if err != nil {
		return http.StatusInternalServerError, []byte("")
	}
	if authBackend.Authenticate(requestUser, &dbUser) {
		token, err := authBackend.GenerateToken(requestUser.Username)
		if err != nil {
			return http.StatusInternalServerError, []byte("")
		} else {
			response, _ := json.Marshal(TokenAuthentication{token})
			return http.StatusOK, response
		}
	}

	return http.StatusUnauthorized, []byte("")
}

func RefreshToken(requestUser *auth.User) []byte {
	authBackend := InitJWTAuthenticationBackend()
	token, err := authBackend.GenerateToken(requestUser.Username)
	if err != nil {
		panic(err)
	}
	response, err := json.Marshal(TokenAuthentication{token})
	if err != nil {
		panic(err)
	}
	return response
}

func Logout(req *http.Request) error {
	authBackend := InitJWTAuthenticationBackend()
	tokenRequest, err := jwtRequst.ParseFromRequest(
		req,
		jwtRequst.AuthorizationHeaderExtractor,
		func(token *jwt.Token) (interface{}, error) {
			return authBackend.PublicKey, nil
		},
	)
	if err != nil {
		return err
	}
	tokenString := req.Header.Get("Authorization")
	return authBackend.Logout(tokenString, tokenRequest)
}

func getPrivateKey() *rsa.PrivateKey {
	privateKeyFile, err := os.Open(settings.PrivateKeyPath)
	if err != nil {
		panic(err)
	}
	defer privateKeyFile.Close()

	pemfileinfo, _ := privateKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(privateKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	privateKeyImported, err := x509.ParsePKCS1PrivateKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	return privateKeyImported
}

func getPublicKey() *rsa.PublicKey {
	publicKeyFile, err := os.Open(settings.PublicKeyPath)
	if err != nil {
		panic(err)
	}

	pemfileinfo, _ := publicKeyFile.Stat()
	var size int64 = pemfileinfo.Size()
	pembytes := make([]byte, size)

	buffer := bufio.NewReader(publicKeyFile)
	_, err = buffer.Read(pembytes)

	data, _ := pem.Decode([]byte(pembytes))

	publicKeyFile.Close()

	publicKeyImported, err := x509.ParsePKIXPublicKey(data.Bytes)

	if err != nil {
		panic(err)
	}

	rsaPub, ok := publicKeyImported.(*rsa.PublicKey)

	if !ok {
		panic(err)
	}

	return rsaPub
}
