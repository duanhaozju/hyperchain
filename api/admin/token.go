//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

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
		return nil, ErrTokenInvalid
	}

	// Is token invalid?
	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	// get claims.
	if claims, err := getJSONFromClaims(token.Claims); err != nil {
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
