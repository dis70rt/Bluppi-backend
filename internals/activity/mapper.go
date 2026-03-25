package activity

import (
    "database/sql"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/friends_activity"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (h *GrpcHandler) mapActivityToProto(a *Activity) *pb.FriendActivity {
    pbAct := &pb.FriendActivity{
        FriendId:        a.FriendID,
        FriendName:      a.FriendName,
        FriendAvatarUrl: a.FriendAvatarURL,
        Status:          a.Status,
    }

    // Because of 'optional' in the proto, these are pointer assignments
    if a.TrackID != nil {
        pbAct.TrackId = a.TrackID
    }
    if a.TrackTitle != nil {
        pbAct.TrackTitle = a.TrackTitle
    }
    if a.TrackArtist != nil {
        pbAct.TrackArtist = a.TrackArtist
    }
    if a.TrackCoverURL != nil {
        pbAct.TrackCoverUrl = a.TrackCoverURL
    }

    return pbAct
}

func (h *GrpcHandler) mapError(err error) error {
    switch err {
    case ErrInvalidInput:
        return status.Error(codes.InvalidArgument, err.Error())
    case sql.ErrNoRows:
        return status.Error(codes.NotFound, "resource not found")
    default:
        // Do not expose raw database metrics to frontend!
        return status.Error(codes.Internal, "internal server error")
    }
}