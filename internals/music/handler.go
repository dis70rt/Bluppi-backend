package music

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/tracks"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
	"github.com/redis/go-redis/v9"
)

type GrpcHandler struct {
    pb.UnimplementedTrackServiceServer
    service *Service
    redis   *redis.Client
}

func NewGrpcHandler(s *Service, redisClient *redis.Client) *GrpcHandler {
    return &GrpcHandler{service: s, redis: redisClient}
}

// --- Core Track Reading ---

func (h *GrpcHandler) GetTrack(ctx context.Context, req *pb.GetTrackRequest) (*pb.TrackResponse, error) {
    track, err := h.service.GetTrack(ctx, req.TrackId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.TrackResponse{
        Track: h.mapTrackToProto(track),
    }, nil
}

// --- Search & Discovery ---

func (h *GrpcHandler) SearchTracks(ctx context.Context, req *pb.SearchTracksRequest) (*pb.SearchTracksResponse, error) {
    tracks, next_cursor, err := h.service.SearchTracks(ctx, req.Query, int(req.Limit), req.Cursor)
    if err != nil {
        return nil, h.mapError(err)
    }

    pbTracks := make([]*pb.SearchTrack, len(tracks))
    for i := range tracks {
        pbTracks[i] = h.mapSearchTrackToProto(&tracks[i])
    }

    return &pb.SearchTracksResponse{
        Tracks: pbTracks,
        Query:  req.Query,
        Limit:  req.Limit,
        NextCursor: next_cursor,
    }, nil
}

func (h *GrpcHandler) GetPopularTracks(ctx context.Context, req *pb.GetPopularTracksRequest) (*pb.GetTracksResponse, error) {
    tracks, err := h.service.GetPopularTracks(ctx, int(req.Limit))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbTracks := make([]*pb.Track, len(tracks))
    for i := range tracks {
        pbTracks[i] = h.mapTrackToProto(&tracks[i])
    }

    return &pb.GetTracksResponse{
        Tracks: pbTracks,
        Total:  int64(len(tracks)),
    }, nil
}

func (h *GrpcHandler) GetTracksByGenre(ctx context.Context, req *pb.GetTracksByGenreRequest) (*pb.GetTracksResponse, error) {
    tracks, total, err := h.service.GetTracksByGenre(ctx, req.Genre, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbTracks := make([]*pb.Track, len(tracks))
    for i := range tracks {
        pbTracks[i] = h.mapTrackToProto(&tracks[i])
    }

    return &pb.GetTracksResponse{
        Tracks: pbTracks,
        Total:  int64(total),
    }, nil
}

// --- User Interactions ---

func (h *GrpcHandler) LikeTrack(ctx context.Context, req *pb.LikeTrackRequest) (*pb.StatusResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }

    if err := h.service.LikeTrack(ctx, userID, req.TrackId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track liked successfully",
    }, nil
}

func (h *GrpcHandler) UnlikeTrack(ctx context.Context, req *pb.UnlikeTrackRequest) (*pb.StatusResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
    
    if err := h.service.UnlikeTrack(ctx, userID, req.TrackId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track unliked successfully",
    }, nil
}

func (h *GrpcHandler) IsTrackLiked(ctx context.Context, req *pb.IsTrackLikedRequest) (*pb.IsTrackLikedResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
    
    liked, err := h.service.IsTrackLiked(ctx, userID, req.TrackId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.IsTrackLikedResponse{
        IsLiked: liked,
    }, nil
}

func (h *GrpcHandler) GetLikedTracks(ctx context.Context, req *pb.GetLikedTracksRequest) (*pb.GetLikedTracksResponse, error) {
    // userID, err := middlewares.GetUserID(ctx)
    // if err != nil {
    //     return nil, h.mapError(err)
    // }
    
    entries, nextCursor, total, err := h.service.GetLikedTracks(ctx, req.TargetUserId, req.Cursor, int(req.Limit))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbEntries := make([]*pb.LikedTrackEntry, len(entries))
    for i := range entries {
        pbEntries[i] = h.mapLikedEntryToProto(&entries[i])
    }

    return &pb.GetLikedTracksResponse{
        Tracks: pbEntries,
        NextCursor: nextCursor,
        Total: int64(total),
        HasMore: nextCursor != "",
    }, nil
}

// --- History ---

func (h *GrpcHandler) AddTrackToHistory(ctx context.Context, req *pb.AddTrackToHistoryRequest) (*pb.StatusResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
    
    if err := h.service.AddTrackToHistory(ctx, userID, req.TrackId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track added to history",
    }, nil
}

func (h *GrpcHandler) GetTrackHistory(ctx context.Context, req *pb.GetTrackHistoryRequest) (*pb.GetTrackHistoryResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
    
    entries, total, err := h.service.GetTrackHistory(ctx, userID, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbEntries := make([]*pb.HistoryTrackEntry, len(entries))
    for i := range entries {
        pbEntries[i] = h.mapHistoryEntryToProto(&entries[i])
    }

    return &pb.GetTrackHistoryResponse{
        History: pbEntries,
        Total:   int64(total),
    }, nil
}

func (h *GrpcHandler) ClearTrackHistory(ctx context.Context, req *pb.ClearTrackHistoryRequest) (*pb.StatusResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
    
    if err := h.service.ClearTrackHistory(ctx, userID); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track history cleared",
    }, nil
}

func (h *GrpcHandler) WeeklyDiscoverTracks(ctx context.Context, req *pb.DiscoverTracksRequest) (*pb.DiscoverTracksResponse, error) {
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }

    tracks, err := h.service.WeeklyDiscoverTracks(ctx, userID, int(req.Limit))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbTracks := make([]*pb.TrackSummary, len(tracks))
    for i := range tracks {
        pbTracks[i] = h.mapTrackToSummaryProto(&tracks[i])
    }

    return &pb.DiscoverTracksResponse{
        Tracks:  pbTracks,
        HasMore: false, 
    }, nil
}

// --- Top Tracks (Profile) ---

func (h *GrpcHandler) GetUserTopTracks(ctx context.Context, req *pb.GetUserTopTracksRequest) (*pb.GetUserTopTracksResponse, error) {
    targetUserID := req.TargetUserId
    if targetUserID == "" {
        return nil, h.mapError(ErrInvalidInput)
    }

    timeRange := req.TimeRange.String()
    limit := int(req.Limit)
    if limit <= 0 || limit > 50 {
        limit = 3
    }

    // --- Redis cache lookup ---
    cacheKey := fmt.Sprintf("top_tracks:%s:%s:%d", targetUserID, timeRange, limit)

    cached, err := h.redis.Get(ctx, cacheKey).Bytes()
    if err == nil {
        var cachedResp pb.GetUserTopTracksResponse
        var entries []TopTrackEntry
        if json.Unmarshal(cached, &entries) == nil {
            pbEntries := make([]*pb.TopTrackEntry, len(entries))
            for i := range entries {
                pbEntries[i] = h.mapTopTrackEntryToProto(&entries[i])
            }
            cachedResp.Tracks = pbEntries
            return &cachedResp, nil
        }
    }

    // --- Cache miss: query DB ---
    entries, err := h.service.GetUserTopTracks(ctx, targetUserID, timeRange, limit)
    if err != nil {
        return nil, h.mapError(err)
    }

    // Cache the result for 12 hours
    if data, err := json.Marshal(entries); err == nil {
        h.redis.Set(ctx, cacheKey, data, 12*time.Hour)
    }

    pbEntries := make([]*pb.TopTrackEntry, len(entries))
    for i := range entries {
        pbEntries[i] = h.mapTopTrackEntryToProto(&entries[i])
    }

    return &pb.GetUserTopTracksResponse{
        Tracks: pbEntries,
    }, nil
}