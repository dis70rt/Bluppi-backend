package users

import (
    "database/sql"

    pb "github.com/dis70rt/bluppi-backend/internals/gen"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func (h *GrpcHandler) mapUserToProto(u *User) *pb.User {
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

func (h *GrpcHandler) mapSearchResultToProto(u *UserSearchResult) *pb.User {
    return &pb.User{
        Id:            u.ID,
        Username:      u.Username,
        Name:          u.Name,
        ProfilePic:    ptrToString(u.ProfilePic),
        FollowerCount: int32(u.FollowerCount),
    }
}

func (h *GrpcHandler) mapRecentSearchToProto(s *RecentSearch) *pb.RecentSearchEntry {
    return &pb.RecentSearchEntry{
        Query:      s.Query,
        SearchedAt: timestamppb.New(s.SearchedAt),
    }
}

func (h *GrpcHandler) mapFollowEntryToProto(f *FollowEntry) *pb.FollowUserEntry {
    return &pb.FollowUserEntry{
        Id:         f.ID,
        Username:   f.Username,
        Name:       f.Name,
        ProfilePic: ptrToString(f.ProfilePic),
        FollowedAt: timestamppb.New(f.FollowedAt),
    }
}

func (h *GrpcHandler) mapError(err error) error {
    switch err {
    case ErrUserNotFound:
        return status.Error(codes.NotFound, err.Error())
    case ErrUserExists:
        return status.Error(codes.AlreadyExists, err.Error())
    case ErrInvalidInput:
        return status.Error(codes.InvalidArgument, err.Error())
    case sql.ErrNoRows:
        return status.Error(codes.NotFound, "resource not found")
    default:
        return status.Error(codes.Internal, err.Error())
    }
}

func ptrToString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}