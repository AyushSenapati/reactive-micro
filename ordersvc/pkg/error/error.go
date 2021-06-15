package error

import "errors"

var (
	// ErrWrongCred should be used when client enters wrong credentials
	ErrWrongCred = errors.New("wrong credentials")

	// ErrTokenExpired should be used whenever the system detects expired token(access/refresh)
	ErrTokenExpired = errors.New("token has expired")

	// ErrApplication should be used in case server failed to process any request and
	// sensitive err info are intended to be kept hidden from the client.
	ErrApplication = errors.New("server encountered some issue, please contact the Admin")

	ErrInsufficientPerm = errors.New("insufficient permission")

	// ErrInvalidReqBody should be used when request body
	// does not match expected fields
	ErrInvalidReqBody = errors.New("invalid request body")
)
