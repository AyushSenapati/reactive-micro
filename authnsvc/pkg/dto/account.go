package dto

type CreateAccountRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	Role     string `json:"role,omitempty"`
}

type CreateAccountResponse struct {
	UserID uint  `json:"user_id,omitempty"`
	Err    error `json:"error,omitempty"`
}

func (resp CreateAccountResponse) Failed() error {
	return resp.Err
}

type GetAccountResponse struct {
	ID           uint   `json:"account_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	TotalRecords int    `json:"-"`
}

type Page struct {
	Page         int `json:"current_page"`
	PageSize     int `json:"page_size"`
	TotalRecords int `json:"total_records"`
}

type ListAccountResponse struct {
	Accounts []GetAccountResponse `json:"accounts,omitempty"`
	PageInfo *Page                `json:"page,omitempty"`
	Err      error                `json:"error,omitempty"`
}

func (resp ListAccountResponse) Failed() error {
	return resp.Err
}
