package jwt

import (
	"context"
	"errors"
	"strings"

	log "github.com/asyrafduyshart/go-exec-engine/pkg/log"
	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

type Jwks struct {
	Keys []Keys `json:"keys"`
}

type Keys struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func ValidateAuth(auth string, jwksUrl string) error {
	tokenString, err := validateBearer(auth)
	if err != nil {
		log.Error("Error validate berarer token: %v", err)
		return err
	}
	_, err = verify(tokenString, jwksUrl)
	if err != nil {
		log.Error("err while parse: %v", err)

	}
	return err
}

func validateBearer(auth string) (string, error) {

	token := strings.Split(auth, "Bearer ")

	if len(token) != 2 {
		return "", errors.New("bearer token nnot in proper format")
	}
	return token[1], nil
}

func verify(tokenString string, url string) (interface{}, error) {
	keySet, err := jwk.Fetch(context.Background(), url)
	if err != nil {
		log.Error("failed to parse JWK: %s", err)
		return nil, err
	}

	tkn, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		kid, status := token.Header["kid"].(string)
		if !status {
			log.Error("kid header not found")
			return nil, errors.New("kid header not found")
		}
		keys, status := keySet.LookupKeyID(kid)
		if !status {
			log.Error("key with specified kid is not present in jwks")
			return nil, errors.New("key with specified kid is not present in jwks")
		}
		var publickey interface{}
		err := keys.Raw(&publickey)
		if err != nil {
			log.Error("could not parse pubkey: %v", err)
			return nil, err
		}
		return publickey, nil
	})
	return tkn, err
}
