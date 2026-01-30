package users

import (
    "context"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/users"
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

func (h *GrpcHandler) GetUserById(ctx context.Context, req *pb.GetUserByIdRequest) (*pb.UserResponse, error) {
    user, err := h.service.GetUserByID(ctx, req.UserId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.UserResponse{
        User: h.mapUserToProto(user),
    }, nil
}

func (h *GrpcHandler) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.UserResponse, error) {
    user, err := h.service.GetUserByUsername(ctx, req.Username)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.UserResponse{
        User: h.mapUserToProto(user),
    }, nil
}

func (h *GrpcHandler) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UserResponse, error) {
    fields := make(map[string]any)

    if req.Email != nil {
        fields["email"] = *req.Email
    }
    if req.Name != nil {
        fields["name"] = *req.Name
    }
    if req.Bio != nil {
        fields["bio"] = *req.Bio
    }
    if req.Country != nil {
        fields["country"] = *req.Country
    }
    if req.Phone != nil {
        fields["phone"] = *req.Phone
    }
    if req.ProfilePic != nil {
        fields["profile_pic"] = *req.ProfilePic
    }
    if len(req.FavoriteGenres) > 0 {
        fields["favorite_genres"] = req.FavoriteGenres
    }

    err := h.service.UpdateUser(ctx, req.UserId, fields)
    if err != nil {
        return nil, h.mapError(err)
    }

    user, err := h.service.GetUserByID(ctx, req.UserId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.UserResponse{
        User: h.mapUserToProto(user),
    }, nil
}

func (h *GrpcHandler) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserResponse, error) {
    err := h.service.DeleteUser(ctx, req.UserId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.DeleteUserResponse{
        Message: "user deleted successfully",
    }, nil
}

func (h *GrpcHandler) SearchUsers(ctx context.Context, req *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error) {
    users, total, err := h.service.SearchUsers(ctx, req.Query, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbUsers := make([]*pb.User, len(users))
    for i, u := range users {
        pbUsers[i] = h.mapSearchResultToProto(&u)
    }

    return &pb.SearchUsersResponse{
        Users: pbUsers,
        Total: int64(total),
    }, nil
}

func (h *GrpcHandler) CheckUsername(ctx context.Context, req *pb.CheckUsernameRequest) (*pb.CheckExistenceResponse, error) {
    exists, err := h.service.UsernameExists(ctx, req.Username)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.CheckExistenceResponse{
        Exists: exists,
    }, nil
}

func (h *GrpcHandler) CheckEmail(ctx context.Context, req *pb.CheckEmailRequest) (*pb.CheckExistenceResponse, error) {
    exists, err := h.service.EmailExists(ctx, req.Email)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.CheckExistenceResponse{
        Exists: exists,
    }, nil
}

func (h *GrpcHandler) GetUserStats(ctx context.Context, req *pb.GetUserStatsRequest) (*pb.UserStatsResponse, error) {
    stats, err := h.service.GetUserStats(ctx, req.UserId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.UserStatsResponse{
        FollowerCount:  int32(stats.FollowersCount),
        FollowingCount: int32(stats.FollowingCount),
    }, nil
}

func (h *GrpcHandler) AddRecentSearch(ctx context.Context, req *pb.AddRecentSearchRequest) (*pb.StatusResponse, error) {
    err := h.service.AddRecentSearch(ctx, req.UserId, req.Query)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Message: "search added",
        Success: true,
    }, nil
}

func (h *GrpcHandler) GetRecentSearches(ctx context.Context, req *pb.GetRecentSearchesRequest) (*pb.RecentSearchesResponse, error) {
    searches, err := h.service.GetRecentSearches(ctx, req.UserId, int(req.Limit))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbSearches := make([]*pb.RecentSearchEntry, len(searches))
    for i, s := range searches {
        pbSearches[i] = h.mapRecentSearchToProto(&s)
    }

    return &pb.RecentSearchesResponse{
        Searches: pbSearches,
    }, nil
}

func (h *GrpcHandler) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.StatusResponse, error) {
    err := h.service.Follow(ctx, req.FollowerId, req.FolloweeId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Message: "user followed",
        Success: true,
    }, nil
}

func (h *GrpcHandler) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.StatusResponse, error) {
    err := h.service.Unfollow(ctx, req.FollowerId, req.FolloweeId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Message: "user unfollowed",
        Success: true,
    }, nil
}

func (h *GrpcHandler) GetFollowers(ctx context.Context, req *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
    followers, total, err := h.service.GetFollowers(ctx, req.UserId, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbFollowers := make([]*pb.FollowUserEntry, len(followers))
    for i, f := range followers {
        pbFollowers[i] = h.mapFollowEntryToProto(&f)
    }

    return &pb.GetFollowersResponse{
        Followers: pbFollowers,
        Total:     int64(total),
    }, nil
}

func (h *GrpcHandler) GetFollowing(ctx context.Context, req *pb.GetFollowingRequest) (*pb.GetFollowingResponse, error) {
    following, total, err := h.service.GetFollowing(ctx, req.UserId, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbFollowing := make([]*pb.FollowUserEntry, len(following))
    for i, f := range following {
        pbFollowing[i] = h.mapFollowEntryToProto(&f)
    }

    return &pb.GetFollowingResponse{
        Following: pbFollowing,
        Total:     int64(total),
    }, nil
}

func (h *GrpcHandler) IsFollowing(ctx context.Context, req *pb.IsFollowingRequest) (*pb.IsFollowingResponse, error) {
    isFollowing, err := h.service.IsFollowing(ctx, req.FollowerId, req.FolloweeId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.IsFollowingResponse{
        IsFollowing: isFollowing,
    }, nil
}