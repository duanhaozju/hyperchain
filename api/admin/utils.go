//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// permissionSet is used to maintain user permissions
type permissionSet map[int]bool

// valid_user records the current valid users with its password.
var valid_user = map[string]string{
	"root": "hyperchain",
}

// user_scope records the current user with its permissions.
var user_scope = map[string]permissionSet{
	"root": rootScopes(),
}

//user_opTimer records the current user with its last operation time.
var user_opTime = map[string]int64{
	"root": 0,
}

// isUserExist checks the username in valid_user, if username is not existed,
// returns ErrUserNotExist; if username is existed but password is not correct,
// returns ErrUnMatch; else returns true.
func isUserExist(username, password string) (bool, error) {
	for k, v := range valid_user {
		if k == username {
			if v == password {
				return true, nil
			} else {
				return false, ErrUnMatch
			}
		}
	}
	return false, ErrUserNotExist
}

// isUserPermit returns if the user has the permission with the given scope.
func isUserPermit(username string, scope int) bool {
	for name, pSet := range user_scope {
		if name == username {
			return pSet[scope]
		}
	}
	return false
}

// basicAuth decodes username and password from http request and returns
// an error if encountered an error.
func basicAuth(req *http.Request) (string, string, error) {
	var username, password string
	auth := req.Header.Get("Authorization")
	if auth == "" {
		log.Debug("Need authorization field in http head.")
		return "", "", ErrNotLogin
	}
	auths := strings.SplitN(auth, " ", 2)
	if len(auths) != 2 {
		log.Debug("Login params need to be split with blank.")
		return "", "", ErrNotLogin
	}
	authMethod := auths[0]
	authB64 := auths[1]
	switch authMethod {
	case "Basic":
		authstr, err := base64.StdEncoding.DecodeString(authB64)
		if err != nil {
			log.Debug(err)
			return "", "", ErrDecodeErr
		}
		userPwd := strings.SplitN(string(authstr), ":", 2)
		if len(userPwd) != 2 {
			log.Debug("username and password must be split by ':'")
			return "", "", ErrDecodeErr
		}
		username = userPwd[0]
		password = userPwd[1]
		return username, password, nil

	default:
		log.Notice("Not supported auth method")
		return "", "", ErrNotSupport
	}
}

// signToken creates, signs, and outputs a token.
func signToken(username, keypath, algorithm string) (string, error) {
	// create a token
	var claims jwt.MapClaims
	tokData := []byte("{}")
	if err := json.Unmarshal(tokData, &claims); err != nil {
		return "", fmt.Errorf("Couldn't parse claims JSON: %v", err)
	}
	claims["iss"] = "Hyperchain Client"
	claims["aud"] = "www.hyperchain.cn"
	claims["usr"] = username

	var key interface{}
	key, err := loadData(keypath)
	if err != nil {
		return "", fmt.Errorf("Couldn't read key: %v", err)
	}

	// get the signing algorithm.
	alg := jwt.GetSigningMethod(algorithm)

	if alg == nil {
		return "", fmt.Errorf("Couldn't find signing method: %v", algorithm)
	}

	// create a new token.
	token := jwt.NewWithClaims(alg, claims)

	if isEs(algorithm) {
		if k, ok := key.([]byte); !ok {
			return "", fmt.Errorf("Couldn't convert key data to key")
		} else {
			key, err = jwt.ParseECPrivateKeyFromPEM(k)
			if err != nil {
				return "", err
			}
		}
	} else if isRs(algorithm) {
		if k, ok := key.([]byte); !ok {
			return "", fmt.Errorf("Couldn't convert key data to key")
		} else {
			key, err = jwt.ParseRSAPrivateKeyFromPEM(k)
			if err != nil {
				return "", err
			}
		}
	}

	return token.SignedString(key)
}

// verifyToken Verifies a token and output the claims.
func verifyToken(auth, keypath, algorithm string) ([]byte, error) {
	// get the token.
	tokData := []byte(auth)

	// trim possible whitespace from token.
	tokData = regexp.MustCompile(`\s*$`).ReplaceAll(tokData, []byte{})

	// Parse the token. Load the key.
	token, err := jwt.Parse(string(tokData), func(t *jwt.Token) (interface{}, error) {
		data, err := loadData(keypath)
		if err != nil {
			log.Debugf("Error in load key data from: %v", keypath)
			return nil, ErrInternal
		}
		if isEs(algorithm) {
			return jwt.ParseECPublicKeyFromPEM(data)
		} else if isRs(algorithm) {
			return jwt.ParseRSAPublicKeyFromPEM(data)
		}
		return data, nil
	})

	// Print an error if we can't parse for some reason.
	if err != nil {
		log.Debugf("Couldn't parse token %s, error: %v", auth, err)
		return nil, ErrTokenInvalid
	}

	// Is token invalid?
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// get claims.
	if claims, err := getJSONFromClaims(token.Claims); err != nil {
		log.Debugf("Failed to parase claims: %v", err)
		return nil, ErrInternal
	} else {
		return claims, nil
	}
}

// loadData reads input from specified file.
func loadData(p string) ([]byte, error) {
	if p == "" {
		return nil, fmt.Errorf("No path specified")
	}
	var reader io.Reader

	if f, err := os.Open(p); err == nil {
		reader = f
		defer f.Close()
	} else {
		return nil, err
	}
	return ioutil.ReadAll(reader)
}

func isEs(alg string) bool {
	return strings.HasPrefix(alg, "ES")
}

func isRs(alg string) bool {
	return strings.HasPrefix(alg, "RS")
}

// getJSONFromClaims returns a json string from claims.
func getJSONFromClaims(j interface{}) ([]byte, error) {
	var out []byte
	var err error

	out, err = json.Marshal(j)

	if err == nil {
		return out, nil
	} else {
		return nil, err
	}
}

// getUserFromClaim returns the username parsed from input claims.
func getUserFromClaim(input []byte) string {
	var claims jwt.MapClaims
	if err := json.Unmarshal(input, &claims); err != nil {
		return ""
	}
	if usr, ok := claims["usr"].(string); !ok {
		return ""
	} else {
		return usr
	}
}

// updateLastOperationTime updates the last operation time.
func updateLastOperationTime(username string) {
	user_opTime[username] = time.Now().Unix()
}