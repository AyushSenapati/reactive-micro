package dto

import stdjwt "github.com/dgrijalva/jwt-go"

// CustomClaim defines the claim to be used in JWT
type CustomClaim struct {
	*stdjwt.StandardClaims
	AccntID uint   `json:"accnt_id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
}

// BasicQueryParam should be used to param basic pagination and orderby queryparams
type BasicQueryParam struct {
	Paginator struct {
		Page     int
		PageSize int
	}
	Filter struct {
		OrederBy []string
	}
}

func NewBasicQueryParam() *BasicQueryParam {
	return &BasicQueryParam{
		Paginator: struct {
			Page     int
			PageSize int
		}{},
		Filter: struct{ OrederBy []string }{},
	}
}
