package music

import (
	"context"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/tracks"
	"github.com/dis70rt/bluppi-backend/internals/infrastructure/middlewares"
)

type GrpcHandler struct {
    pb.UnimplementedTrackServiceServer
    service *Service
}

func NewGrpcHandler(s *Service) *GrpcHandler {
    return &GrpcHandler{service: s}
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
    userID, err := middlewares.GetUserID(ctx)
    if err != nil {
        return nil, h.mapError(err)
    }
    
    entries, total, err := h.service.GetLikedTracks(ctx, userID, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbEntries := make([]*pb.LikedTrackEntry, len(entries))
    for i := range entries {
        pbEntries[i] = h.mapLikedEntryToProto(&entries[i])
    }

    return &pb.GetLikedTracksResponse{
        Tracks: pbEntries,
        Total:  int64(total),
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