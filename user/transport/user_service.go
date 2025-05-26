package transport

import (
	"context"
	"strconv"
	"user/internal/service"
	"user/pkg/userpb"

	"gorm.io/gorm"
)

type UserServiceServer struct {
	userpb.UnimplementedUserServiceServer
	userService *service.UserService
}

func NewUserServiceServer(db *gorm.DB) *UserServiceServer {
	return &UserServiceServer{
		userService: service.NewUserService(db),
	}
}

func (s *UserServiceServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	userID64, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, err
	}
	userID := uint(userID64)

	exists, err := s.userService.CheckByID(userID)
	if err != nil {
		return nil, err
	}

	return &userpb.GetUserResponse{
		Exists: exists,
	}, nil
}
