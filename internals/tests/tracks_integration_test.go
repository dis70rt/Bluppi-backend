package tests

import (
    "context"
    "net"
    "os"
    "testing"
    "time"

    "github.com/dis70rt/bluppi-backend/internals/music"
    "github.com/dis70rt/bluppi-backend/internals/users"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func setupMusicTests(t *testing.T, db *sqlx.DB) (*music.Service, *users.Service, func()) {
    t.Helper()

    tx, err := db.Beginx()
    require.NoError(t, err)

    userRepo := users.NewRepositoryWithTx(tx)
    userService := users.NewService(userRepo)

    musicRepo := music.NewRepositoryWithTx(tx)

    grpcAddr := os.Getenv("GRPC_SEARCH_SERVICE_ADDR")
    if grpcAddr == "" {
        grpcAddr = "localhost:50052"
    }

    musicService := music.NewService(musicRepo, grpcAddr)

    cleanup := func() {
        _ = tx.Rollback()
    }

    return musicService, userService, cleanup
}

func createTestTrack(id, title, artist string) *music.Track {
    return &music.Track{
        ID:         id,
        Title:      title,
        Artist:     artist,
        Album:      StringPtr("Test Album"),
        Duration:   180,
        Genre:      []string{"Pop", "Rock"},
        Listeners:  100,
        PlayCount:  500,
        Popularity: 50,
        CreatedAt:  time.Now(),
    }
}

// Helper to generate a valid UUID for tests
func newUUID() string {
    return uuid.New().String()
}

// ==================== CreateTrack Tests ====================

func TestCreateTrack_Success(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    // Use valid UUID
    track := createTestTrack(newUUID(), "Song A", "Artist A")

    err := service.CreateTrack(ctx, track)

    assert.NoError(t, err)
}

func TestCreateTrack_InvalidInput(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()

    err := service.CreateTrack(ctx, nil)
    assert.ErrorIs(t, err, music.ErrInvalidInput)

    emptyTrack := &music.Track{}
    err = service.CreateTrack(ctx, emptyTrack)
    assert.ErrorIs(t, err, music.ErrInvalidInput)
}

// ==================== GetTrack Tests ====================

func TestGetTrack_Success(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    id := newUUID()
    track := createTestTrack(id, "Get Song", "Get Artist")
    require.NoError(t, service.CreateTrack(ctx, track))

    result, err := service.GetTrack(ctx, id)

    assert.NoError(t, err)
    assert.Equal(t, track.ID, result.ID)
    assert.Equal(t, track.Title, result.Title)
}

func TestGetTrack_NotFound(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()

    // Use valid UUID that doesn't exist
    _, err := service.GetTrack(ctx, newUUID())
    assert.ErrorIs(t, err, music.ErrTrackNotFound)
}

// ==================== UpdateTrack Tests ====================

func TestUpdateTrack_Success(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    id := newUUID()
    track := createTestTrack(id, "Old Title", "Old Artist")
    require.NoError(t, service.CreateTrack(ctx, track))

    updates := map[string]any{
        "title":      "New Title",
        "popularity": 99,
    }

    err := service.UpdateTrack(ctx, id, updates)
    assert.NoError(t, err)

    updated, err := service.GetTrack(ctx, id)
    assert.NoError(t, err)
    assert.Equal(t, "New Title", updated.Title)
    assert.Equal(t, 99, updated.Popularity)
}

// ==================== DeleteTrack Tests ====================

func TestDeleteTrack_Success(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    id := newUUID()
    track := createTestTrack(id, "Delete Me", "Artist D")
    require.NoError(t, service.CreateTrack(ctx, track))

    err := service.DeleteTrack(ctx, id)
    assert.NoError(t, err)

    _, err = service.GetTrack(ctx, id)
    assert.ErrorIs(t, err, music.ErrTrackNotFound)
}

// ==================== Popular & Genre Tests ====================

func TestGetPopularTracks(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    t1 := createTestTrack(newUUID(), "Hit Song", "Artist")
    t1.Popularity = 100
    t2 := createTestTrack(newUUID(), "Okay Song", "Artist")
    t2.Popularity = 50

    require.NoError(t, service.CreateTrack(ctx, t1))
    require.NoError(t, service.CreateTrack(ctx, t2))

    tracks, err := service.GetPopularTracks(ctx, 10)

    assert.NoError(t, err)
    assert.GreaterOrEqual(t, len(tracks), 2)
}

func TestGetTracksByGenre(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    t1 := createTestTrack(newUUID(), "Rock Anthem", "Rocker")
    t1.Genre = []string{"Rock", "Metal"}
    t2 := createTestTrack(newUUID(), "Jazz Smooth", "Jazzer")
    t2.Genre = []string{"Jazz"}

    require.NoError(t, service.CreateTrack(ctx, t1))
    require.NoError(t, service.CreateTrack(ctx, t2))

    tracks, total, err := service.GetTracksByGenre(ctx, "Rock", 10, 0)
    assert.NoError(t, err)
    assert.GreaterOrEqual(t, total, 1)
    assert.GreaterOrEqual(t, len(tracks), 1)

    for _, tr := range tracks {
        contains := false
        for _, g := range tr.Genre {
            if g == "Rock" {
                contains = true
                break
            }
        }
        assert.True(t, contains, "Track should contain requested genre")
    }
}

// ==================== Search Integration Tests ====================

func TestSearchTracks(t *testing.T) {
    grpcAddr := os.Getenv("GRPC_SEARCH_SERVICE_ADDR")
    if grpcAddr == "" {
        grpcAddr = "localhost:50052"
    }

    // Skip if service is not running
    conn, err := net.DialTimeout("tcp", grpcAddr, 100*time.Millisecond)
    if err != nil {
        t.Skipf("Skipping integration test: gRPC service not running at %s", grpcAddr)
    }
    conn.Close()

    db := GetTestDB(t)
    defer db.Close()

    service, _, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    query := "Eminem"

    results, total, err := service.SearchTracks(ctx, query, 10, 0)

    assert.NoError(t, err, "Search service should not return error")
    if total > 0 {
        assert.NotEmpty(t, results)
        assert.NotEmpty(t, results[0].Title)
    }
}

// ==================== Like System Tests ====================

func TestLikeTrack_Success(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    mService, uService, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()

    // Use valid UUIDs for User and Track
    user := createTestUser(newUUID(), "liker_"+newUUID()[:8], "like"+newUUID()+"@test.com")
    require.NoError(t, uService.CreateUser(ctx, user))
    track := createTestTrack(newUUID(), "Liked Song", "Artist")
    require.NoError(t, mService.CreateTrack(ctx, track))

    err := mService.LikeTrack(ctx, user.ID, track.ID)
    assert.NoError(t, err)

    liked, err := mService.IsTrackLiked(ctx, user.ID, track.ID)
    assert.NoError(t, err)
    assert.True(t, liked)
}

func TestUnlikeTrack(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    mService, uService, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    user := createTestUser(newUUID(), "unliker_"+newUUID()[:8], "unlike"+newUUID()+"@test.com")
    require.NoError(t, uService.CreateUser(ctx, user))
    track := createTestTrack(newUUID(), "Song", "Art")
    require.NoError(t, mService.CreateTrack(ctx, track))

    require.NoError(t, mService.LikeTrack(ctx, user.ID, track.ID))

    err := mService.UnlikeTrack(ctx, user.ID, track.ID)
    assert.NoError(t, err)

    liked, err := mService.IsTrackLiked(ctx, user.ID, track.ID)
    assert.NoError(t, err)
    assert.False(t, liked)
}

func TestGetLikedTracks(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    mService, uService, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    user := createTestUser(newUUID(), "listlikes_"+newUUID()[:8], "list"+newUUID()+"@test.com")
    require.NoError(t, uService.CreateUser(ctx, user))

    for i := 0; i < 3; i++ {
        track := createTestTrack(newUUID(), "S", "A") // Valid UUID for each track
        require.NoError(t, mService.CreateTrack(ctx, track))
        require.NoError(t, mService.LikeTrack(ctx, user.ID, track.ID))
    }

    likes, total, err := mService.GetLikedTracks(ctx, user.ID, 10, 0)
    assert.NoError(t, err)
    assert.Equal(t, 3, total)
    assert.Len(t, likes, 3)
}

// ==================== History Tests ====================

func TestHistory_Flow(t *testing.T) {
    db := GetTestDB(t)
    defer db.Close()

    mService, uService, cleanup := setupMusicTests(t, db)
    defer cleanup()

    ctx := context.Background()
    user := createTestUser(newUUID(), "historyuser_"+newUUID()[:8], "hist"+newUUID()+"@test.com")
    require.NoError(t, uService.CreateUser(ctx, user))
    track := createTestTrack(newUUID(), "History Song", "Artist")
    require.NoError(t, mService.CreateTrack(ctx, track))

    err := mService.AddTrackToHistory(ctx, user.ID, track.ID)
    assert.NoError(t, err)

    hist, total, err := mService.GetTrackHistory(ctx, user.ID, 10, 0)
    assert.NoError(t, err)
    assert.Equal(t, 1, total)
    assert.Equal(t, track.ID, hist[0].TrackID)

    err = mService.ClearTrackHistory(ctx, user.ID)
    assert.NoError(t, err)

    histAfter, totalAfter, err := mService.GetTrackHistory(ctx, user.ID, 10, 0)
    assert.NoError(t, err)
    assert.Equal(t, 0, totalAfter)
    assert.Len(t, histAfter, 0)
}