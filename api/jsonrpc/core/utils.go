//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"math/big"
	"reflect"
	"unicode"
	"unicode/utf8"
	"strings"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"io"
	"os"
	"io/ioutil"
	"net/http"
	"time"
	"github.com/dgrijalva/jwt-go"
	"encoding/json"
	"strings"
	"errors"
)

// permissionSet used to maintain user permissions
type permissionSet map[int]bool

var valid_user = map [string]string {
	"root"    : "hyperchain",
	"duanhao" : "123",
}

var user_scope = map [string]permissionSet {
	"root"    : rootScopes(),
	"duanhao" : defaultScopes(),
}

func splitRawMessage(args json.RawMessage) ([]string, error) {
	str := string(args[:])
	if len(str) < 4 {
		return nil, errors.New("invalid args")
	}
	str = str[2 : len(str)-2]
	splitstr := strings.Split(str, ",")
	return splitstr, nil
}

var bigIntType = reflect.TypeOf((*big.Int)(nil)).Elem()

// Indication if this type should be serialized in hex
func isHexNum(t reflect.Type) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t == bigIntType
}


func IsUserExist(username, password string) (bool, error) {
	for k,v := range valid_user {
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

func IsUserPermit(username string, scope int) bool {
	for name, scopes := range user_scope {
		if name == username{
			return contains(scopes, scope)
		}
	}
	return false
}

func contains(pset permissionSet, scope int) bool {
	return pset[scope]
}

func basicAuth(req *http.Request) (string, string, error){
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

// Create, sign, and output a token.
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
	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Unix() + expiration
	claims["nbf"] = time.Now().Unix() - beforetime
	// Do we need to wrap the scope to the claims?
	//if scope, ok := user_scope[username]; !ok {
	//	return "", fmt.Errorf("Couldn't find the scope of user: %s", username)
	//} else {
	//	claims["scp"] = scope
	//}

	// get the key
	var key interface{}
	key, err := loadData(keypath)
	if err != nil {
		return "", fmt.Errorf("Couldn't read key: %v", err)
	}

	// get the signing alg
	alg := jwt.GetSigningMethod(algorithm)

	if alg == nil {
		return "", fmt.Errorf("Couldn't find signing method: %v", algorithm)
	}

	// create a new token
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

// Verify a token and output the claims.
func verifyToken(auth, keypath, algorithm string) ([]byte, error) {
	// get the token
	tokData := []byte(auth)

	// trim possible whitespace from token
	tokData = regexp.MustCompile(`\s*$`).ReplaceAll(tokData, []byte{})

	// Parse the token. Load the key
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

	// Print an error if we can't parse for some reason
	if err != nil {
		log.Debugf("Couldn't parse token %s, error: %v", auth, err)
		return nil, ErrTokenInvalid
	}

	// Is token invalid?
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// get claims
	if claims, err := getJSONFromClaims(token.Claims); err != nil {
		log.Debugf("Failed to parase claims: %v", err)
		return nil, ErrInternal
	} else {
		return claims, nil
	}
}

// loadData reads input from specified file
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

// getJSONFromClaims returns a json string from claims
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

// checkPermissionByToken checks if method is permitted in input claims
func checkPermissionByToken(input []byte, method string) (bool, error) {
	scope := convertToScope(method)
	if scope == -1 {
		return false, ErrPermission
	}

	claims := string(input)

	// find scp field in claims
	pat := `"scp":\[.+\]`
	reg, err := regexp.Compile(pat)
	if (err != nil ) {
		log.Debug(err)
		return false, ErrInternal
	}
	matchStr := reg.FindString(claims)
	if matchStr == "" {
		return false, ErrPermission
	}
	//scopeStr := matchStr[6:]

	//var scopes []int
	//json.Unmarshal([]byte(scopeStr), &scopes)
	//if contains(scopes, scope) {
	//	return true, nil
	//}
	return false, ErrPermission
}

// checkPermission checks permission by username in input claims
func checkPermission(input []byte, method string) (bool, error) {
	scope := convertToScope(method)
	if scope == -1 {
		return false, ErrPermission
	}

	var claims jwt.MapClaims
	if err := json.Unmarshal(input, &claims); err != nil {
		return false, ErrInternal
	}
	if usr, ok := claims["usr"].(string); !ok {
		return false, ErrInternal
	} else {
		if IsUserPermit(usr, scope) {
			return true, nil
		} else {
			return false, ErrPermission
		}
	}
}

// createUser creates a new account with given username and password
func createUser(username, password, group string) error {
	groupPermission := getGroupPermission(group)
	if groupPermission == nil {
		return ErrInvalidGroup
	}
	valid_user[username] = password
	user_scope[username] = groupPermission
	return nil
}

// alterUser alters an existed account with given username and password
func alterUser(username, password string) {
	valid_user[username] = password
}

// delUser deletes an account
func delUser(username string) {
	delete(valid_user, username)
	delete(user_scope, username)
}
