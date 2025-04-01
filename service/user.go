package service

import (
	"chinese-chess-backend/database"
	dto "chinese-chess-backend/dto/user"
	userModel "chinese-chess-backend/model/user"
	"chinese-chess-backend/utils"
	"time"

	"errors"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (us *UserService) Register(req *dto.RegisterRequest) (dto.RegisterResponse, error) {
	var registerResp dto.RegisterResponse
	var err error

	user := userModel.User{
		Name:     req.Name,
		Email:    req.Email,
	}

	// 校验邮箱是否已注册
	db := database.GetMysqlDb()
	if err = db.Where("email = ?", req.Email).First(&user).Error; err == nil {
		return registerResp, errors.New("邮箱已注册")
	}

	code, err := database.GetValue(req.Email)

	if err != nil {
		return registerResp, err
	}

	if code != req.VCode {
		err = errors.New("验证码错误")
		return registerResp, err
	}

	hashPass, err := utils.HashPassword(req.Password)
	if err != nil {
		return registerResp, err
	}

	user.Password = hashPass

	db.Create(&user)

	// 签发token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return registerResp, err
	}

	registerResp.Token = token
	registerResp.Name = user.Name

	return registerResp, nil
}

func (us *UserService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	var loginResp dto.LoginResponse
	var err error

	// 查询用户
	db := database.GetMysqlDb()
	user := userModel.User{}
	if err = db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// 校验密码
	if !utils.CheckPassword(user.Password, req.Password) {
		return nil, errors.New("密码错误")
	}

	// 签发token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return nil, err
	}

	loginResp.Token = token
	loginResp.Name = user.Name

	return &loginResp, nil
}

func (us *UserService) SendVCode(req *dto.SendVCodeRequest) error {
	var err error

	// 生成验证码
	code := utils.GetRandomCode()

	// 发送验证码
	err = utils.SendMail(req.Email, "验证码", code)
	if err != nil {
		return err
	}

	// 存储验证码到数据库
	err = database.SetValue(req.Email, code, time.Minute * 5)
	if err != nil {
		return err
	}

	return nil
}

func (uc *UserService) GetUserInfo(req *dto.GetUserInfoRequest) (*dto.GetUserInfoResponse, error) {
	var userInfoResp dto.GetUserInfoResponse
	var err error

	// 查询用户
	db := database.GetMysqlDb()
	user := userModel.User{}
	if err = db.Where("id = ?", req.Id).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	userInfoResp.Name = user.Name

	return &userInfoResp, nil
}