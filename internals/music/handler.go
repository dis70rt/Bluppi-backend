package music

import (
    "context"

    pb "github.com/dis70rt/bluppi-backend/internals/gen/tracks"
)

type GrpcHandler struct {
    pb.UnimplementedTrackServiceServer
    service *Service
}

func NewGrpcHandler(s *Service) *GrpcHandler {
    return &GrpcHandler{service: s}
}

func (h *GrpcHandler) CreateTrack(ctx context.Context, req *pb.CreateTrackRequest) (*pb.TrackResponse, error) {
    track := &Track{
        ID:         req.Id,
        Title:      req.Title,
        Artist:     req.Artist,
        Album:      stringToPtr(req.Album),
        Duration:   int(req.Duration),
        Genre:      req.Genres,
        ImageURL:   stringToPtr(req.ImageUrl),
        PreviewURL: stringToPtr(req.PreviewUrl),
        VideoID:    stringToPtr(req.VideoId),
        Listeners:  int(req.Listeners),
        PlayCount:  int(req.PlayCount),
        Popularity: int(req.Popularity),
    }

    if err := h.service.CreateTrack(ctx, track); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.TrackResponse{
        Track: h.mapTrackToProto(track),
    }, nil
}

func (h *GrpcHandler) GetTrack(ctx context.Context, req *pb.GetTrackRequest) (*pb.TrackResponse, error) {
    track, err := h.service.GetTrack(ctx, req.TrackId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.TrackResponse{
        Track: h.mapTrackToProto(track),
    }, nil
}

func (h *GrpcHandler) UpdateTrack(ctx context.Context, req *pb.UpdateTrackRequest) (*pb.TrackResponse, error) {
    fields := make(map[string]any)

    if req.Title != nil {
        fields["title"] = *req.Title
    }
    if req.Artist != nil {
        fields["artist"] = *req.Artist
    }
    if req.Album != nil {
        fields["album"] = *req.Album
    }
    if req.Duration != nil {
        fields["duration"] = int(*req.Duration)
    }
    if len(req.Genres) > 0 {
        fields["genre"] = req.Genres
    }
    if req.ImageUrl != nil {
        fields["image_url"] = *req.ImageUrl
    }
    if req.PreviewUrl != nil {
        fields["preview_url"] = *req.PreviewUrl
    }
    if req.VideoId != nil {
        fields["video_id"] = *req.VideoId
    }
    if req.Listeners != nil {
        fields["listeners"] = int(*req.Listeners)
    }
    if req.PlayCount != nil {
        fields["play_count"] = int(*req.PlayCount)
    }
    if req.Popularity != nil {
        fields["popularity"] = int(*req.Popularity)
    }

    if err := h.service.UpdateTrack(ctx, req.TrackId, fields); err != nil {
        return nil, h.mapError(err)
    }

    track, err := h.service.GetTrack(ctx, req.TrackId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.TrackResponse{
        Track: h.mapTrackToProto(track),
    }, nil
}

func (h *GrpcHandler) DeleteTrack(ctx context.Context, req *pb.DeleteTrackRequest) (*pb.DeleteTrackResponse, error) {
    if err := h.service.DeleteTrack(ctx, req.TrackId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.DeleteTrackResponse{
        Message: "track deleted successfully",
    }, nil
}

func (h *GrpcHandler) SearchTracks(ctx context.Context, req *pb.SearchTracksRequest) (*pb.SearchTracksResponse, error) {
    tracks, total, err := h.service.SearchTracks(ctx, req.Query, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, h.mapError(err)
    }

    pbTracks := make([]*pb.Track, len(tracks))
    for i := range tracks {
        pbTracks[i] = h.mapTrackToProto(&tracks[i])
    }

    return &pb.SearchTracksResponse{
        Tracks: pbTracks,
        Total:  int64(total),
        Query:  req.Query,
        Limit:  req.Limit,
        Offset: req.Offset,
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

func (h *GrpcHandler) LikeTrack(ctx context.Context, req *pb.LikeTrackRequest) (*pb.StatusResponse, error) {
    if err := h.service.LikeTrack(ctx, req.UserId, req.TrackId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track liked successfully",
    }, nil
}

func (h *GrpcHandler) UnlikeTrack(ctx context.Context, req *pb.UnlikeTrackRequest) (*pb.StatusResponse, error) {
    if err := h.service.UnlikeTrack(ctx, req.UserId, req.TrackId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track unliked successfully",
    }, nil
}

func (h *GrpcHandler) IsTrackLiked(ctx context.Context, req *pb.IsTrackLikedRequest) (*pb.IsTrackLikedResponse, error) {
    liked, err := h.service.IsTrackLiked(ctx, req.UserId, req.TrackId)
    if err != nil {
        return nil, h.mapError(err)
    }

    return &pb.IsTrackLikedResponse{
        IsLiked: liked,
    }, nil
}

func (h *GrpcHandler) GetLikedTracks(ctx context.Context, req *pb.GetLikedTracksRequest) (*pb.GetLikedTracksResponse, error) {
    entries, total, err := h.service.GetLikedTracks(ctx, req.UserId, int(req.Limit), int(req.Offset))
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

func (h *GrpcHandler) AddTrackToHistory(ctx context.Context, req *pb.AddTrackToHistoryRequest) (*pb.StatusResponse, error) {
    if err := h.service.AddTrackToHistory(ctx, req.UserId, req.TrackId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track added to history",
    }, nil
}

func (h *GrpcHandler) GetTrackHistory(ctx context.Context, req *pb.GetTrackHistoryRequest) (*pb.GetTrackHistoryResponse, error) {
    entries, total, err := h.service.GetTrackHistory(ctx, req.UserId, int(req.Limit), int(req.Offset))
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
    if err := h.service.ClearTrackHistory(ctx, req.UserId); err != nil {
        return nil, h.mapError(err)
    }

    return &pb.StatusResponse{
        Success: true,
        Message: "track history cleared",
    }, nil
}