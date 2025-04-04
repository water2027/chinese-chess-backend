package user

import (
	"fmt"
)

type GetUserInfoRequest struct {
	Id int `json:"id"`
}

func (r *GetUserInfoRequest) Examine() error {
	if r.Id <= 0 {
		return fmt.Errorf("用户ID无效")
	}
	return nil
}

type GetUserInfoResponse struct {
	UserInfo
}