package models

type LoginForm struct {
	Email    string `json:"email" binding:"required" example:"username@email.com"`
	Password string `json:"password" binding:"required" example:"X7B2K9M4P1Q6R3L8A5C0J2N7D4F9E6H3"`
}

type Authenticator func(form LoginForm) (a *Account, err error)
