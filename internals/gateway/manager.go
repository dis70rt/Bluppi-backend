package gateway

import (
    "sync"

    "github.com/google/uuid"
)

type Connection struct {
    ID                 string
    UserID             string
    Chan               chan PresenceEvent
    SubscribedToTarget []string // which users this connection is watching
}

type ConnectionManager struct {
    mu sync.RWMutex
    // userId -> connectionId -> Connection
    userConnections map[string]map[string]*Connection
    // targetUserId -> set of connectionIds watching them
    targetWatchers map[string]map[string]struct{}
    // connectionId -> Connection reference (for quick lookup during push)
    connById map[string]*Connection
}

func NewConnectionManager() *ConnectionManager {
    return &ConnectionManager{
        userConnections: make(map[string]map[string]*Connection),
        targetWatchers:  make(map[string]map[string]struct{}),
        connById:        make(map[string]*Connection),
    }
}

func (cm *ConnectionManager) AddConnection(userID string, targetUserIDs []string) (*Connection, string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    connID := uuid.New().String()
    conn := &Connection{
        ID:                 connID,
        UserID:             userID,
        Chan:               make(chan PresenceEvent, 20), // slight buffer increase for safety
        SubscribedToTarget: targetUserIDs,
    }

    // Register user connection
    if cm.userConnections[userID] == nil {
        cm.userConnections[userID] = make(map[string]*Connection)
    }
    cm.userConnections[userID][connID] = conn
    cm.connById[connID] = conn

    // Register watchers
    for _, targetID := range targetUserIDs {
        if cm.targetWatchers[targetID] == nil {
            cm.targetWatchers[targetID] = make(map[string]struct{})
        }
        cm.targetWatchers[targetID][connID] = struct{}{}
    }

    return conn, connID
}

func (cm *ConnectionManager) RemoveConnection(userID, connID string) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    conn, exists := cm.connById[connID]
    if !exists {
        return
    }

    // Unregister from watchers
    for _, targetID := range conn.SubscribedToTarget {
        if watchers := cm.targetWatchers[targetID]; watchers != nil {
            delete(watchers, connID)
            if len(watchers) == 0 {
                delete(cm.targetWatchers, targetID)
            }
        }
    }

    // Clean up user connections mapping
    if userConns := cm.userConnections[userID]; userConns != nil {
        delete(userConns, connID)
        if len(userConns) == 0 {
            delete(cm.userConnections, userID)
        }
    }

    // Clean up global mappings and close channel
    delete(cm.connById, connID)
    close(conn.Chan)
}

// PushEvent broadcasts an event to anyone who subscribed to the event.UserID
func (cm *ConnectionManager) PushEvent(targetUserID string, event PresenceEvent) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    watchers, exists := cm.targetWatchers[targetUserID]
    if !exists {
        return
    }

    // Send to every connection watching this target
    for connID := range watchers {
        if conn, ok := cm.connById[connID]; ok {
            select {
            case conn.Chan <- event:
            default: // Drop event if channel is full to avoid blocking or Deadlock
            }
        }
    }
}