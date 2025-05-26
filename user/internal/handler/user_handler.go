package handler

import "user/internal/service"

type UserHandler struct {
	s *service.UserService
}

func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{s: s}
}
