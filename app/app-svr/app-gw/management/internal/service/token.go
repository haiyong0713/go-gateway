package service

import (
	"context"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-gw/management/internal/model"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

func (s *CommonService) newToken(req *model.JWTTokenPayload, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, req)
	return token.SignedString(secret)
}

func (s *CommonService) parseToken(tokenStr string, secret []byte) (*model.JWTTokenPayload, error) {
	payload := &model.JWTTokenPayload{}
	token, err := jwt.ParseWithClaims(tokenStr, payload, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("Unexpected signing method: %+v", token.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse jwt: %s", tokenStr)
	}
	if !token.Valid {
		return nil, errors.New("Invalid token: " + tokenStr)
	}
	claims := token.Claims.(*model.JWTTokenPayload)
	return claims, nil
}

func (s *CommonService) initialTokenSecret(ctx context.Context) {
	if err := s.dao.InitialTokenSecret(ctx); err != nil {
		log.Error("Failed to initial token secret: %+v", err)
	}
}
