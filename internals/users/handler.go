package users

import (
	"context"

	pb "github.com/dis70rt/bluppi-backend/internals/gen"
	"github.com/dis70rt/bluppi-backend/internals/utils"
)

type GrpcHandler struct {
	pb.UnimplementedUserServiceServer
	service *Service
}

func NewGrpcHandler(s *Service) *GrpcHandler {
	return &GrpcHandler{service: s}
}

func (h *GrpcHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	user := &User{
		ID:             req.Id,
		Email:          req.Email,
		Username:       req.Username,
		Name:           req.Name,
		Bio:            utils.StringToPtr(req.Bio),
		Country:        utils.StringToPtr(req.Country),
		Phone:          utils.StringToPtr(req.Phone),
		ProfilePic:     utils.StringToPtr(req.ProfilePic),
		FavoriteGenres: req.FavoriteGenres,
	}

	err := h.service.CreateUser(ctx, user)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &pb.UserResponse{
		User: h.mapUserToProto(user),
	}, nil
}