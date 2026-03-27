package music

import (
    "database/sql"
    "strings"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/tracks"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/timestamppb"
)

// mapTrackToProto converts the internal Domain Track to the Proto message
func (h *GrpcHandler) mapTrackToProto(t *Track) *pb.Track {
    if t == nil {
        return nil
    }

    // Handle genres: stored as TEXT in DB (e.g., "Pop, Rock"), repeated in Proto
    var genres []string
    if t.Genres != "" {
        if strings.Contains(t.Genres, ",") {
            parts := strings.Split(t.Genres, ",")
            for _, p := range parts {
                genres = append(genres, strings.TrimSpace(p))
            }
        } else {
            genres = []string{t.Genres}
        }
    }

    return &pb.Track{
        Id:         t.ID,
        Title:      t.Title,
        Artist:     t.Artists, // Mapped to 'Artist' field in Proto
        DurationMs: int32(t.DurationMS),
        Genres:     genres,
        ImageSmall: ptrToString(t.ImageSmall),
        ImageLarge: ptrToString(t.ImageLarge),
        PreviewUrl: ptrToString(t.PreviewURL),
        VideoId:    ptrToString(t.VideoID),
        Listeners:  t.Listeners,
        PlayCount:  t.PlayCount,
        Popularity: int32(t.Popularity),
        CreatedAt:  timestamppb.New(t.CreatedAt),
    }
}

func (h *GrpcHandler) mapLikedEntryToProto(e *LikedTrackEntry) *pb.LikedTrackEntry {
    if e == nil {
        return nil
    }

    return &pb.LikedTrackEntry{
        TrackId:    e.TrackID,
        Title:      e.Title,
        Artist:     e.Artists, // Mapped to 'Artist' field in Proto
        ImageSmall: ptrToString(e.ImageSmall),
        LikedAt:    timestamppb.New(e.LikedAt),
    }
}

func (h *GrpcHandler) mapHistoryEntryToProto(e *HistoryEntry) *pb.HistoryTrackEntry {
    if e == nil {
        return nil
    }

    return &pb.HistoryTrackEntry{
        TrackId:    e.TrackID,
        Title:      e.Title,
        Artist:     e.Artists, // Mapped to 'Artist' field in Proto
        ImageSmall: ptrToString(e.ImageSmall),
        PlayedAt:   timestamppb.New(e.PlayedAt),
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

func (h *GrpcHandler) mapSearchTrackToProto(t *SearchTrack) *pb.SearchTrack {
    if t == nil {
        return nil
    }
    return &pb.SearchTrack{
        Id:         t.ID,
        Title:      t.Title,
        Artist:     t.Artists, // Map to 'artist' field in proto
        ImageSmall: ptrToString(t.ImageSmall),
        PreviewUrl: ptrToString(t.PreviewURL),
    }
}

func (h *GrpcHandler) mapTrackToSummaryProto(t *Track) *pb.TrackSummary {
    if t == nil {
        return nil
    }
    
    return &pb.TrackSummary{
        Id:         t.ID,
        Title:      t.Title,
        Artist:     t.Artists,
        ImageUrl:   ptrToString(t.ImageLarge),
        PreviewUrl: ptrToString(t.PreviewURL),
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