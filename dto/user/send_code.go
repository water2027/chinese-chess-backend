package user

import (
	"fmt"
	"github.com/dlclark/regexp2"
)

type SendVCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (r *SendVCodeRequest) Examine() error {	
	reg := regexp2.MustCompile(`^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$`, 0)
	isMatch, _ := reg.MatchString(r.Email)
	if !isMatch {
		return fmt.Errorf("邮箱不合格")
	}

	return nil
}