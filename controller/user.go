package controller

import (
	"github.com/gin-gonic/gin"

	"chinese-chess-backend/service"
	"chinese-chess-backend/dto"
	"chinese-chess-backend/dto/user"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) Register(c *gin.Context) {
	var req user.RegisterRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	resp, err := uc.userService.Register(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))

}

func (uc *UserController) Login(c *gin.Context) {
	var req user.LoginRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	resp, err := uc.userService.Login(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}

func (uc *UserController) SendVCode(c *gin.Context) {
	var req user.SendVCodeRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	err = uc.userService.SendVCode(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithMessage("验证码发送成功"))
}

func (uc *UserController) GetUserInfo(c *gin.Context) {
	var req user.GetUserInfoRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	resp, err := uc.userService.GetUserInfo(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}