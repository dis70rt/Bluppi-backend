package music

import (
    "database/sql"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/tracks"
    ytpb "github.com/dis70rt/bluppi-backend/internals/gen/ytmusic"
    "github.com/lib/pq"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

func MapProtoTrackToDomain(t *ytpb.Track) Track {
    if t == nil {
        return Track{}
    }

    track := Track{
        ID:         t.TrackId,
        Title:      t.TrackName,
        Artist:     t.ArtistName,
        Duration:   int(t.GetDuration()),
        Listeners:  int(t.Listeners),
        PlayCount:  int(t.Playcount),
        Popularity: int(t.Popularity),
        Genre:      pq.StringArray(t.Genres),
    }

    track.Album = t.AlbumName
    track.ImageURL = t.ImageUrl
    track.PreviewURL = t.PreviewUrl

    if t.VideoId != "" {
        vid := t.VideoId
        track.VideoID = &vid
    }

    return track
}

func MapProtoTracksToDomain(protoTracks []*ytpb.Track) []Track {
    if protoTracks == nil {
        return []Track{}
    }

    tracks := make([]Track, 0, len(protoTracks))
    for _, t := range protoTracks {
        tracks = append(tracks, MapProtoTrackToDomain(t))
    }
    return tracks
}

func (h *GrpcHandler) mapTrackToProto(t *Track) *pb.Track {
    if t == nil {
        return nil
    }

    return &pb.Track{
        Id:         t.ID,
        Title:      t.Title,
        Artist:     t.Artist,
        Album:      ptrToString(t.Album),
        Duration:   int32(t.Duration),
        Genres:     []string(t.Genre),
        ImageUrl:   ptrToString(t.ImageURL),
        PreviewUrl: ptrToString(t.PreviewURL),
        VideoId:    ptrToString(t.VideoID),
        Listeners:  int64(t.Listeners),
        PlayCount:  int64(t.PlayCount),
        Popularity: int32(t.Popularity),
        CreatedAt:  timestamppb.New(t.CreatedAt),
    }
}

func (h *GrpcHandler) mapTrackToSummary(t *Track) *pb.TrackSummary {
    if t == nil {
        return nil
    }

    return &pb.TrackSummary{
        Id:       t.ID,
        Title:    t.Title,
        Artist:   t.Artist,
        ImageUrl: ptrToString(t.ImageURL),
        Duration: int32(t.Duration),
    }
}

func (h *GrpcHandler) mapLikedEntryToProto(e *LikedTrackEntry) *pb.LikedTrackEntry {
    if e == nil {
        return nil
    }

    return &pb.LikedTrackEntry{
        TrackId:  e.TrackID,
        Title:    e.Title,
        Artist:   e.Artist,
        Album:    ptrToString(e.Album),
        ImageUrl: ptrToString(e.ImageURL),
        LikedAt:  timestamppb.New(e.LikedAt),
    }
}

func (h *GrpcHandler) mapHistoryEntryToProto(e *HistoryEntry) *pb.HistoryTrackEntry {
    if e == nil {
        return nil
    }

    return &pb.HistoryTrackEntry{
        TrackId:  e.TrackID,
        Title:    e.Title,
        Artist:   e.Artist,
        Album:    ptrToString(e.Album),
        ImageUrl: ptrToString(e.ImageURL),
        PlayedAt: timestamppb.New(e.PlayedAt),
    }
}

// ----------------- Error Mapping -----------------

func (h *GrpcHandler) mapError(err error) error {
    switch err {
    case ErrTrackNotFound:
        return status.Error(codes.NotFound, err.Error())
    case ErrInvalidInput:
        return status.Error(codes.InvalidArgument, err.Error())
    case ErrAlreadyLiked:
        return status.Error(codes.AlreadyExists, err.Error())
    case ErrNotLiked:
        return status.Error(codes.NotFound, err.Error())
    case ErrHistoryEmpty:
        return status.Error(codes.NotFound, err.Error())
    case sql.ErrNoRows:
        return status.Error(codes.NotFound, "resource not found")
    default:
        return status.Error(codes.Internal, err.Error())
    }
}

// ----------------- Helper Functions -----------------

func ptrToString(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}

func stringToPtr(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}