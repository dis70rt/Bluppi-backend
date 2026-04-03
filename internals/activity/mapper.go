package activity

import (
    "database/sql"
    "time"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/friends_activity" // ensure correct path
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func (h *GrpcHandler) mapActivityToProto(a *Activity) *pb.FriendActivity {
    pbAct := &pb.FriendActivity{
        FriendId:        a.FriendID,
        FriendName:      a.FriendName,
        FriendAvatarUrl: a.FriendAvatarURL,
        Status:          a.Status,
        FriendUsername:  a.FriendUsername,
    }

    if a.LastSeen > 0 {
        pbAct.LastSeen = timestamppb.New(time.Unix(a.LastSeen, 0))
    } else {
        pbAct.LastSeen = timestamppb.New(time.Now())
    }

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
    if a.TrackPreviewURL != nil {
        pbAct.TrackPreviewUrl = a.TrackPreviewURL
    }

    return pbAct
}

func mapError(err error) error {
    switch err {
    case sql.ErrNoRows:
        return status.Error(codes.NotFound, "data not found")
    case ErrInvalidInput:
        return status.Error(codes.InvalidArgument, "invalid argument provided")
    default:
        return status.Error(codes.Internal, "internal server error: "+err.Error())
    }
}