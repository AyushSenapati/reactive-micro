package dto

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Err          error  `json:"err,omitempty"`
}

func (resp LoginResponse) Failed() error {
	return resp.Err
}
