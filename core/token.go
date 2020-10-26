package core

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	uuid "github.com/satori/go.uuid"
)

func EncodeAccessToken(id uuid.UUID, subject string, duration *time.Duration) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))
	issAt := time.Now()
	t.Claims = &Claims{
		StandardClaims: &jwt.StandardClaims{
			Id:       id.String(),
			IssuedAt: issAt.Unix(),
			Issuer:   "ONE RA",
			Subject:  subject,
		},
	}
	if duration != nil {
		t.Claims.(*Claims).ExpiresAt = issAt.Add(*duration).Unix()
	}
	return t.SignedString(signKey.Key)
}

func DecodeAccessToken(t string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(t, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return &signKey.Key.PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("decode token error")
}
