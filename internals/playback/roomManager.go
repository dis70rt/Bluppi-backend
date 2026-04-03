package playback

import (
	"log"
	"sync"
	"time"

	pb "github.com/dis70rt/bluppi-backend/internals/gen/playback"
)

type RoomManager struct {
	mu   sync.RWMutex
	rooms map[string]*RoomState
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*RoomState),
	}
}

func (m *RoomManager) GetRoom(hostID, roomID string) *RoomState {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[roomID]
	if !exists {
		room = &RoomState{
			ID: roomID,
			HostUserID: hostID,
			Clients: make(map[string]*ClientSession),
			BufferReady: make(map[string]bool),
			State: PlaybackState{
				Version: 1,
			},
		}
		m.rooms[roomID] = room
	}

	return room
}

func (m *RoomManager) CleanupRoom(roomID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rooms, roomID)
	log.Printf("Room %s destroyed from memory", roomID)
}

func (r *RoomState) RegisterClient(c *ClientSession) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Clients[c.UserID] = c
}

func (r *RoomState) UnregisterClient(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.Clients, userID)
	delete(r.BufferReady, userID)
}

func (r *RoomState) SendCurrentState(c *ClientSession) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c.Send <- r.buildStateEvent()
}

func (r *RoomState) Broadcast() {
	r.mu.RLock()
	event := r.buildStateEvent()
	for _ , client := range r.Clients {
		select {
		case client.Send <- event:
		default:
			log.Printf("Warning: Client %s stream is blocked", client.UserID)
		}
	}
	r.mu.RUnlock()
}

func (r *RoomState) HandleTrackChange(cmd *pb.TrackChangeCommand) {
	r.mu.Lock()

	r.State.Track = TrackInfo{
		ID: cmd.TrackId,
		Title: cmd.Title,
		Artist: cmd.Artist,
		AudioURL: cmd.AudioUrl,
		ArtworkURL: cmd.ArtworkUrl,
		DurationMs: cmd.DurationMs,
	}

	r.State.IsPlaying = false
	r.State.PositionMs = 0
	r.State.ScheduledStartServerUs = 0
	r.State.Version++
	r.BufferReady = make(map[string]bool)

	r.mu.Unlock()
	r.Broadcast()
}

func (r *RoomState) HandleBufferReady(userID string, version uint64) {
	r.mu.Lock()

	if version != r.State.Version {
		r.mu.Unlock()
		return
	}

	r.BufferReady[userID] = true

	totalClients := len(r.Clients)
	readyClients := len(r.BufferReady)

	if totalClients == 0 {
		r.mu.Unlock()
		return
	}

	quorumMet := float64(readyClients) / float64(totalClients) >= 0.8

	if !quorumMet || r.State.IsPlaying {
		r.mu.Unlock()
		return
	}

	nowUs := time.Now().UnixMicro()
	r.State.IsPlaying = true
	r.State.LastUpdateServerUs = nowUs
	r.State.ScheduledStartServerUs = nowUs + 500000
	r.State.Version++
	
	r.mu.Unlock()
	r.Broadcast()
}

func (r *RoomState) HandlePlay() {
	r.mu.Lock()
	if r.State.IsPlaying {
		r.mu.Unlock()
		return
	}

	nowUs := time.Now().UnixMicro()
	r.State.IsPlaying = true
	r.State.LastUpdateServerUs = nowUs
	r.State.ScheduledStartServerUs = nowUs + 300000
	r.State.Version++
	
	r.mu.Unlock()
	r.Broadcast()
}

func (r *RoomState) HandlePause() {
	r.mu.Lock()
	if !r.State.IsPlaying {
		r.mu.Unlock()
		return
	}

	nowUs := time.Now().UnixMicro()
	elapsedUs := nowUs - r.State.LastUpdateServerUs
	elapsedMs := elapsedUs / 1000

	r.State.IsPlaying = false
	r.State.PositionMs += elapsedMs
	r.State.LastUpdateServerUs = nowUs
	r.State.ScheduledStartServerUs = 0
	r.State.Version++
	
	r.mu.Unlock()
	r.Broadcast()
}

func (r *RoomState) buildStateEvent() *pb.ServerEvent {
	track := r.State.Track
	return &pb.ServerEvent{
		Payload: &pb.ServerEvent_PlaybackState{
			PlaybackState: &pb.PlaybackState{
				Track:                  &pb.TrackInfo{
					TrackId: track.ID,
					Title: track.Title,
					Artist: track.Artist,
					ArtworkUrl: track.ArtworkURL,
					DurationMs: track.DurationMs,
					AudioUrl: track.AudioURL,
				},
				IsPlaying:              r.State.IsPlaying,
				PositionMs:             r.State.PositionMs,
				LastUpdateServerUs:     r.State.LastUpdateServerUs,
				ScheduledStartServerUs: r.State.ScheduledStartServerUs,
				Version:                r.State.Version,
			},
		},
	}
}