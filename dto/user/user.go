package user

type UserInfo struct {
	ID    uint    `json:"id"`
	Token string `json:"token"`
	Name  string `json:"name"`
	Exp   int    `json:"exp"`
}