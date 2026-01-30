import grpc
from concurrent import futures
import logging
import os
import signal
import sys

from ytmusicService import YTMusicService

import ytmusic_pb2
import ytmusic_pb2_grpc

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class YTMusicGRPCServicer(ytmusic_pb2_grpc.YTMusicServiceServicer):
    
    def __init__(self):
        self.ytmusic_service = YTMusicService()
        logger.info("YTMusicGRPCServicer initialized")
    
    def _dict_to_track(self, track_dict: dict) -> ytmusic_pb2.Track:

        return ytmusic_pb2.Track(
            track_id=track_dict.get("trackId", ""),
            track_name=track_dict.get("trackName", ""),
            artist_name=track_dict.get("artistName", "Unknown Artist"),
            album_name=track_dict.get("albumName"),
            duration=track_dict.get("duration"),
            genres=track_dict.get("genres", []),
            image_url=track_dict.get("imageUrl"),
            preview_url=track_dict.get("previewUrl"),
            video_id=track_dict.get("videoId", ""),
            listeners=track_dict.get("listeners", 0),
            playcount=track_dict.get("playcount", 0),
            popularity=track_dict.get("popularity", 0)
        )
    
    def SearchTracks(self, request: ytmusic_pb2.SearchRequest, 
                     context: grpc.ServicerContext) -> ytmusic_pb2.SearchResponse:
        logger.info(f"SearchTracks called: query='{request.query}', limit={request.limit}, offset={request.offset}")
        
        try:
            limit = request.limit if request.limit > 0 else 20
            offset = request.offset if request.offset >= 0 else 0
            
            result = self.ytmusic_service.search_tracks(
                query=request.query,
                limit=limit,
                offset=offset
            )
            
            tracks = [self._dict_to_track(t) for t in result.get("tracks", [])]
            
            response = ytmusic_pb2.SearchResponse(
                status=result.get("status", "error"),
                status_code=result.get("status_code", 500),
                tracks=tracks,
                total=result.get("total", 0),
                limit=result.get("limit", limit),
                offset=result.get("offset", offset),
                message=result.get("message")
            )
            
            logger.info(f"SearchTracks returning {len(tracks)} tracks")
            return response
            
        except Exception as e:
            logger.error(f"SearchTracks error: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return ytmusic_pb2.SearchResponse(
                status="error",
                status_code=500,
                message=str(e)
            )
    
    def GetRecommendations(self, request: ytmusic_pb2.RecommendationRequest,
                           context: grpc.ServicerContext) -> ytmusic_pb2.RecommendationResponse:
        logger.info(f"GetRecommendations called: video_id='{request.video_id}', limit={request.limit}")
        
        try:
            limit = request.limit if request.limit > 0 else 5
            
            result = self.ytmusic_service.get_recommendations(
                video_id=request.video_id,
                limit=limit
            )
            
            tracks = [self._dict_to_track(t) for t in result.get("tracks", [])]
            
            response = ytmusic_pb2.RecommendationResponse(
                status=result.get("status", "error"),
                status_code=result.get("status_code", 500),
                tracks=tracks,
                total=result.get("total", 0),
                message=result.get("message")
            )
            
            logger.info(f"GetRecommendations returning {len(tracks)} tracks")
            return response
            
        except Exception as e:
            logger.error(f"GetRecommendations error: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return ytmusic_pb2.RecommendationResponse(
                status="error",
                status_code=500,
                message=str(e)
            )


def serve():
    host = os.getenv("GRPC_HOST", "0.0.0.0")
    port = os.getenv("GRPC_PORT", "50052")
    max_workers = int(os.getenv("GRPC_MAX_WORKERS", "10"))
    
    server = grpc.server(
        futures.ThreadPoolExecutor(max_workers=max_workers),
        options=[
            ('grpc.max_send_message_length', 50 * 1024 * 1024),  
            ('grpc.max_receive_message_length', 50 * 1024 * 1024),  
        ]
    )
    
    ytmusic_pb2_grpc.add_YTMusicServiceServicer_to_server(
        YTMusicGRPCServicer(), server
    )
    
    server_address = f"{host}:{port}"
    server.add_insecure_port(server_address)
    
    def shutdown_handler(signum, frame):
        logger.info("Received shutdown signal, stopping server...")
        stopped_event = server.stop(grace=5)
        stopped_event.wait()
        logger.info("Server stopped")
        sys.exit(0)
    
    signal.signal(signal.SIGTERM, shutdown_handler)
    signal.signal(signal.SIGINT, shutdown_handler)
    
    server.start()
    logger.info(f"gRPC server started on {server_address}")
    
    server.wait_for_termination()


if __name__ == "__main__":
    serve()