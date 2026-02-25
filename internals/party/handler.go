package party

import (
	"context"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/party"
)

type GrpcHandler struct {
	service *Service
	pb.UnimplementedSyncServiceServer
}

func NewGrpcHandler(s *Service) *GrpcHandler {
	return &GrpcHandler{service: s}
}

func (h *GrpcHandler) ClockSync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	return &pb.SyncResponse{
		ServerReceiveUs: CaptureServerReceiveUs(),
		ServerSendUs:    CaptureServerSendUs(),
	}, nil
}