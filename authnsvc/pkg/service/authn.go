package service

import (
	"context"
	"fmt"
	"time"

	svcconf "github.com/AyushSenapati/reactive-micro/authnsvc/conf"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/dto"
	ce "github.com/AyushSenapati/reactive-micro/authnsvc/pkg/error"
	"github.com/AyushSenapati/reactive-micro/authnsvc/pkg/util"
	stdjwt "github.com/dgrijalva/jwt-go"
	kitjwt "github.com/go-kit/kit/auth/jwt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (svc *basicAuthNService) GenToken(ctx context.Context, req dto.LoginRequest) dto.LoginResponse {
	accntObj, err := svc.accntrepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return dto.LoginResponse{Err: ce.ErrWrongCred}
		}
		return dto.LoginResponse{Err: err}
	}

	if !util.CheckPasswordHash(req.Password, accntObj.Password) {
		return dto.LoginResponse{Err: ce.ErrWrongCred}
	}

	accessToken, err := svc.genAccessToken(accntObj.ID, accntObj.Email, accntObj.Role.Name)
	if err != nil {
		return dto.LoginResponse{Err: err}
	}

	refreshToken, err := svc.genRefreshToken(accntObj.ID, accntObj.Email, accntObj.Role.Name)

	return dto.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken, Err: err}
}

func (svc *basicAuthNService) newCustomClaim(aid uint, email, role string, ttl time.Duration) dto.CustomClaim {
	claims := dto.CustomClaim{
		AccntID: aid,
		Email:   email,
		Role:    role,
		StandardClaims: &stdjwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(ttl).Unix(),
			Issuer:    svcconf.C.Auth.Issuer,
		},
	}
	return claims
}

func (svc *basicAuthNService) genAccessToken(aid uint, email, role string) (string, error) {
	claims := svc.newCustomClaim(aid, email, role, svcconf.C.Auth.AccessTokenTTL)

	accessToken := stdjwt.NewWithClaims(stdjwt.SigningMethodHS256, claims)
	accessToken.Header["kid"] = svcconf.C.Auth.AccessKID // access token ID
	return accessToken.SignedString([]byte(svcconf.C.Auth.SecretKey))
}

func (svc *basicAuthNService) genRefreshToken(aid uint, email, role string) (string, error) {
	claims := svc.newCustomClaim(aid, email, role, svcconf.C.Auth.RefreshTokenTTL)
	claims.StandardClaims.Id = uuid.New().String() // Adds JTI to the refresh token

	refreshToken := stdjwt.NewWithClaims(stdjwt.SigningMethodHS256, claims)
	refreshToken.Header["kid"] = svcconf.C.Auth.RefreshKID // refresh token ID
	signedRefreshToken, err := refreshToken.SignedString([]byte(svcconf.C.Auth.SecretKey))

	return signedRefreshToken, err
}

// ValidateToken validates the given token based on the provided kid (KeyID)
func (svc *basicAuthNService) ValidateToken(tokenString, kid string) (dto.CustomClaim, error) {
	claim := dto.CustomClaim{}

	kf := func(token *stdjwt.Token) (interface{}, error) {
		keyID := token.Header["kid"].(string)
		if keyID != kid {
			return claim, stdjwt.ErrInvalidKeyType
		}
		return []byte(svcconf.C.Auth.SecretKey), nil
	}

	token, err := stdjwt.ParseWithClaims(tokenString, &claim, kf)

	// check if signature is valid
	if err != nil {
		return claim, err
	}
	if token.Valid {
		return claim, nil
	}
	return claim, kitjwt.ErrTokenInvalid
}

func (svc *basicAuthNService) RenewAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claim, err := svc.ValidateToken(refreshToken, svcconf.C.Auth.RefreshKID)
	if err != nil {
		return "", err
	}
	if svc.authnrepo.IsBlacklisted(ctx, claim.Id) {
		return "", ce.ErrTokenExpired
	}
	accessToken, err := svc.genAccessToken(claim.AccntID, claim.Email, claim.Role)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

// VerifyToken checks if the access token is valid and is allowed to access the services
func (svc *basicAuthNService) VerifyToken(ctx context.Context, accessToken string) bool {
	_, err := svc.ValidateToken(accessToken, svcconf.C.Auth.AccessKID)
	return err == nil
}

func (svc *basicAuthNService) RevokeToken(ctx context.Context, refreshToken string) bool {
	claim, err := svc.ValidateToken(refreshToken, svcconf.C.Auth.RefreshKID)
	if err != nil {
		return false
	}
	if claim.ExpiresAt > time.Now().Unix() {
		timeUnix := time.Unix(claim.ExpiresAt, 0)
		exp := time.Until(timeUnix)
		isBlacklisted, err := svc.authnrepo.Blacklist(ctx, claim.Id, exp)
		if err != nil {
			fmt.Println(err)
		}
		return isBlacklisted
	}
	return true
}
