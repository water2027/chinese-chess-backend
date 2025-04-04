package user

import (
	"fmt"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func (r *LoginRequest) Examine() error {
	email := r.Email
	password := r.Password
	if email == "" || password == "" {
		return fmt.Errorf("邮箱和密码不能为空")
	}
	return nil
}

type LoginResponse struct {
	UserInfo
}