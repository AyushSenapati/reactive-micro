package endpoint

import (
	"github.com/AyushSenapati/reactive-micro/inventorysvc/pkg/dto"
	stdjwt "github.com/dgrijalva/jwt-go"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	kitep "github.com/go-kit/kit/endpoint"
)

func NewJWTTokenParsingMW(secretKey string) kitep.Middleware {
	kf := func(token *stdjwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	}

	claimFactory := func() stdjwt.Claims { return &dto.CustomClaim{} }
	return kitjwt.NewParser(kf, stdjwt.SigningMethodHS256, claimFactory)
}
