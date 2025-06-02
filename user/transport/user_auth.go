package transport

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"user/internal/model"
	"user/internal/service"
	"user/pkg/auth_user_pb"
)

type UserAuthServer struct {
	auth_user_pb.AuthServiceServer
	s *service.UserService
}

func NewUserAuthServer(db *gorm.DB) *UserAuthServer {
	return &UserAuthServer{
		s: service.NewUserService(db),
	}
}

func (serv *UserAuthServer) Register(ctx context.Context, req *auth_user_pb.RegisterRequest) (*auth_user_pb.RegisterResponse, error) {
	user := model.User{
		Username: req.Username,
		Password: req.Password,
	}

	id, err := serv.s.CreateUser(&user)
	if err != nil {
		return &auth_user_pb.RegisterResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &auth_user_pb.RegisterResponse{
		Id:      uint64(id),
		Success: true,
		Error:   "",
	}, nil
}

func (serv *UserAuthServer) Login(ctx context.Context, req *auth_user_pb.LoginRequest) (*auth_user_pb.LoginResponse, error) {
	user, err := serv.s.GetUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &auth_user_pb.LoginResponse{
				Success: false,
				Error:   "user not found",
			}, nil
		}
		return &auth_user_pb.LoginResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	if user.Password != req.Password {
		return &auth_user_pb.LoginResponse{
			Success: false,
			Error:   "invalid password",
		}, nil
	}

	return &auth_user_pb.LoginResponse{
		Id:      uint64(user.ID),
		Success: true,
		Error:   "",
	}, nil
}

func (serv *UserAuthServer) LoginWithGoogle(ctx context.Context, req *auth_user_pb.GoogleLoginRequest) (*auth_user_pb.GoogleLoginResponse, error) {
	user, err := serv.s.GetUserByUsername(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newUser := model.User{
				Username: req.Email,
				Password: "",
			}
			id, createErr := serv.s.CreateUser(&newUser)
			if createErr != nil {
				return &auth_user_pb.GoogleLoginResponse{
					Success: false,
					Error:   createErr.Error(),
				}, nil
			}
			return &auth_user_pb.GoogleLoginResponse{
				Id:      uint64(id),
				Success: true,
				Error:   "",
			}, nil
		}
		return &auth_user_pb.GoogleLoginResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &auth_user_pb.GoogleLoginResponse{
		Id:      uint64(user.ID),
		Success: true,
		Error:   "",
	}, nil
}
