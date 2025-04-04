package user

import (
	"fmt"
	"github.com/dlclark/regexp2"
	"strings"
)

type RegisterRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Password2 string `json:"password2" binding:"required,min=6"`
	VCode     string `json:"vcode" binding:"required"`
}

func (r *RegisterRequest) Examine() error {
	if r.Password != r.Password2 {
		return fmt.Errorf("两次密码输入不一致")
	}
	reg := regexp2.MustCompile(`^\S*(?=\S{6,})(?=\S*\d)(?=\S*[A-Z])(?=\S*[a-z])(?=\S*[!@#$%^&*? ])\S*$`, 0)
	isMatch, _ := reg.MatchString(r.Password)
	if !isMatch {
		return fmt.Errorf("密码必须包含数字、大写字母、小写字母和特殊字符，且长度至少为6位")
	}
	// 去除空格查看是否为空
	// 去除空格
	name := strings.TrimSpace(r.Name)
	email := strings.TrimSpace(r.Email)
	vcode := strings.TrimSpace(r.VCode)
	if name == "" || email == "" || vcode == "" {
		return fmt.Errorf("用户名、邮箱和验证码不能为空")
	}
	if len(name) < 3 || len(name) > 20 {
		return fmt.Errorf("用户名长度必须在3到20个字符之间")
	}
	if len(email) < 5 || len(email) > 50 {
		return fmt.Errorf("邮箱长度必须在5到50个字符之间")
	}

	return nil
}

type RegisterResponse struct {
	UserInfo
}
