package users

import (
	"context"
	"errors"
	
	pb "github.com/dis70rt/bluppi-backend/internals/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GrpcHandler struct {
	pb.UnimplementedUserServiceServer
	service *Service
}

func NewGrpcHandler(s *Service) *GrpcHandler {
	return &GrpcHandler{service: s}
}

func (h *GrpcHandler) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	
	domainUser := &User{
		ID:             req.Id,
		Email:          req.Email,
		Username:       req.Username,
		Name:           req.Name,
		
		Bio:            stringToPtr(req.Bio),
		Country:        stringToPtr(req.Country),
		Phone:          stringToPtr(req.Phone),
		ProfilePic:     stringToPtr(req.ProfilePic),
		FavoriteGenres: req.FavoriteGenres,
	}

	err := h.service.CreateUser(ctx, domainUser)
	if err != nil {
		return nil, h.mapError(err)
	}

	return &pb.UserResponse{User: h.mapUserToProto(domainUser)}, nil
}

func (h *GrpcHandler) mapUserToProto(u *User) *pb.User {
	if u == nil {
		return nil
	}
	return &pb.User{
		Id:             u.ID,
		Email:          u.Email,
		Username:       u.Username,
		Name:           u.Name,
		
		Bio:            ptrToString(u.Bio),
		Country:        ptrToString(u.Country),
		Phone:          ptrToString(u.Phone),
		ProfilePic:     ptrToString(u.ProfilePic),
		FavoriteGenres: u.FavoriteGenres,
		FollowerCount:  int32(u.FollowerCount),
		FollowingCount: int32(u.FollowingCount),
		CreatedAt:      timestamppb.New(u.CreatedAt),
	}
}

func (h *GrpcHandler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input data")
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

func stringToPtr(s string) *string {
	if s == "" {
		return nil 
	}
	return &s
}

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}