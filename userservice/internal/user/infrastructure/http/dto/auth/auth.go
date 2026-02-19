package auth

type SignInReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterReq struct {
	Email    string  `json:"email" validate:"required,email,max=254"`
	Password string  `json:"password" validate:"required,min=8,password"`
	Username string  `json:"username" validate:"required,min=3,max=50,alphanum"`
	Name     string  `json:"name" validate:"required,min=1,max=100"`
	ZipCode  *string `json:"zip_code,omitempty" validate:"omitempty,max=20"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,max=20"`
}

type SignInResp struct {
	UserID       int    `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResp struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}
