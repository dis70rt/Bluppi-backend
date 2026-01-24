package users

import (
	"errors"
	"log"

	pb "github.com/dis70rt/bluppi-backend/internals/gen"
	"github.com/dis70rt/bluppi-backend/internals/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *GrpcHandler) mapUserToProto(u *User) *pb.User {
	if u == nil {
		return nil
	}

	return &pb.User{
		Id:       u.ID,
		Email:    u.Email,
		Username: u.Username,
		Name:     u.Name,

		Bio:        utils.PtrToString(u.Bio),
		Country:    utils.PtrToString(u.Country),
		Phone:      utils.PtrToString(u.Phone),
		ProfilePic: utils.PtrToString(u.ProfilePic),

		FavoriteGenres: u.FavoriteGenres,
		FollowerCount:  int32(u.FollowerCount),
		FollowingCount: int32(u.FollowingCount),

		CreatedAt: timestamppb.New(u.CreatedAt),
	}
}

func (h *GrpcHandler) mapError(err error) error {
	switch {
	case errors.Is(err, ErrUserNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, "user already exists")
	case errors.Is(err, ErrInvalidInput):
		return status.Error(codes.InvalidArgument, "invalid input arguments")
	case errors.Is(err, ErrAlreadyFollowing):
		return status.Error(codes.AlreadyExists, "already following this user")
	case errors.Is(err, ErrNotFollowing):
		return status.Error(codes.FailedPrecondition, "cannot unfollow: not currently following")
	default:
		log.Printf("Internal error: %v", err)
		return status.Error(codes.Internal, "internal server error")
	}
}