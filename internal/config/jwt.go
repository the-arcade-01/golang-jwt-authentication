package config

import "github.com/go-chi/jwtauth/v5"

func newJWTAuthClient() *jwtauth.JWTAuth {
	return jwtauth.New("HS256", []byte(Envs.JWT_SECRET_KEY), nil)
}
