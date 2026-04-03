package tests
// // TODO: Implement genre make the query fast.
// import (
//     "context"
//     "testing"
//     "time"

//     "github.com/dis70rt/bluppi-backend/internals/music"
//     "github.com/dis70rt/bluppi-backend/internals/users"
//     "github.com/google/uuid"
//     "github.com/jmoiron/sqlx"
//     "github.com/stretchr/testify/assert"
//     "github.com/stretchr/testify/require"
// )

// type MusicTestContext struct {
//     MusicService *music.Service
//     UserService  *users.Service
//     TX           *sqlx.Tx
//     Cleanup      func()
// }

// func setupMusicTests(t *testing.T, db *sqlx.DB) *MusicTestContext {
//     t.Helper()

//     tx, err := db.Beginx()
//     require.NoError(t, err)

//     userRepo := users.NewRepositoryWithTx(tx)
//     bus := &noOpPublisher{}
//     userService := users.NewService(userRepo, bus)

//     musicRepo := music.NewRepositoryWithTx(tx, nil)
//     musicService := music.NewService(musicRepo)

//     cleanup := func() {
//         _ = tx.Rollback()
//     }

//     return &MusicTestContext{
//         MusicService: musicService,
//         UserService:  userService,
//         TX:           tx,
//         Cleanup:      cleanup,
//     }
// }

// func newUUID() string {
//     return uuid.New().String()
// }

// func createTestTrackObject(id, title, artists string) *music.Track {
//     return &music.Track{
//         ID:         id,
//         Title:      title,
//         Artists:    artists,
//         Genres:     "Pop, Rock",
//         DurationMS: 180000,
//         Listeners:  100,
//         PlayCount:  500,
//         Popularity: 50,
//         CreatedAt:  time.Now(),
//         ImageSmall: StringPtr("http://img.small"),
//         ImageLarge: StringPtr("http://img.large"),
//     }
// }

// func seedTrack(t *testing.T, tx *sqlx.Tx, tr *music.Track) {
//     t.Helper()

//     query := `
//         INSERT INTO tracks (
//             track_id, title, artists, genres, duration_ms, 
//             popularity, created_at, image_small, image_large, preview_url, video_id
//         ) VALUES (
//             :track_id, :title, :artists, :genres, :duration_ms, 
//             :popularity, :created_at, :image_small, :image_large, :preview_url, :video_id
//         )
//     `
//     _, err := tx.NamedExec(query, tr)
//     require.NoError(t, err, "failed to seed track")

//     statsQuery := `
//         INSERT INTO track_stats (track_id, listeners, play_count) 
//         VALUES ($1, $2, $3)
//     `
//     _, err = tx.Exec(statsQuery, tr.ID, tr.Listeners, tr.PlayCount)
//     require.NoError(t, err, "failed to seed track stats")
// }

// // ==================== GetTrack Tests ====================

// func TestGetTrack_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     ctx := setupMusicTests(t, db)
//     defer ctx.Cleanup()

//     id := newUUID()
//     track := createTestTrackObject(id, "Get Song", "Get Artist")

//     seedTrack(t, ctx.TX, track)

//     result, err := ctx.MusicService.GetTrack(context.Background(), id)

//     assert.NoError(t, err)
//     assert.NotNil(t, result)
//     assert.Equal(t, track.ID, result.ID)
//     assert.Equal(t, track.Title, result.Title)
//     assert.Equal(t, track.Artists, result.Artists)
//     assert.Equal(t, int64(500), result.PlayCount)
// }

// func TestGetTrack_NotFound(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     ctx := setupMusicTests(t, db)
//     defer ctx.Cleanup()

//     _, err := ctx.MusicService.GetTrack(context.Background(), newUUID())
//     assert.ErrorIs(t, err, music.ErrTrackNotFound)
// }

// // ==================== Popular & Genre Tests ====================

// func TestGetPopularTracks(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     ctx := setupMusicTests(t, db)
//     defer ctx.Cleanup()

//     t1 := createTestTrackObject(newUUID(), "Hit Song", "Artist")
//     t1.Popularity = 100
//     t2 := createTestTrackObject(newUUID(), "Okay Song", "Artist")
//     t2.Popularity = 50

//     seedTrack(t, ctx.TX, t1)
//     seedTrack(t, ctx.TX, t2)

//     tracks, err := ctx.MusicService.GetPopularTracks(context.Background(), 10)

//     assert.NoError(t, err)
//     assert.GreaterOrEqual(t, len(tracks), 2)
//     assert.True(t, tracks[0].Popularity >= tracks[1].Popularity)
// }

// // func TestGetTracksByGenre(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     ctx := setupMusicTests(t, db)
// //     defer ctx.Cleanup()

// //     // Use a unique genre that won't exist in production data
// //     uniqueGenre := "TestGenre_" + newUUID()[:8]

// //     t1 := createTestTrackObject(newUUID(), "Rock Anthem", "Rocker")
// //     t1.Genres = uniqueGenre + ", Metal"
// //     t2 := createTestTrackObject(newUUID(), "Jazz Smooth", "Jazzer")
// //     t2.Genres = "Jazz"

// //     seedTrack(t, ctx.TX, t1)
// //     seedTrack(t, ctx.TX, t2)

// //     tracks, total, err := ctx.MusicService.GetTracksByGenre(context.Background(), uniqueGenre, 10, 0)
// //     assert.NoError(t, err)
// //     assert.Equal(t, 1, total)
// //     assert.Len(t, tracks, 1)

// //     found := false
// //     for _, tr := range tracks {
// //         if strings.Contains(tr.Genres, uniqueGenre) {
// //             found = true
// //             break
// //         }
// //     }
// //     assert.True(t, found, "Returned tracks should contain the requested genre")
// // }

// // ==================== Search Tests (Internal FTS) ====================

// //TODO: IMPLEMENT SOLR TESTS. This is currently not possible because the solr client is not mocked and we don't have a test solr instance. We can either mock the solr client or set up a test solr instance for integration testing.
// // func TestSearchTracks(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     ctx := setupMusicTests(t, db)
// //     defer ctx.Cleanup()

// //     uniqueArtist1 := "UniqueArtist_" + newUUID()[:8]
// //     uniqueArtist2 := "AnotherUnique_" + newUUID()[:8]

// //     t1 := createTestTrackObject(newUUID(), "Lose Yourself", uniqueArtist1)
// //     t1.Genres = "Hip-Hop"
// //     t1.Popularity = 100
// //     seedTrack(t, ctx.TX, t1)

// //     t2 := createTestTrackObject(newUUID(), "Shape of You", uniqueArtist2)
// //     t2.Genres = "Pop"
// //     t2.Popularity = 100
// //     seedTrack(t, ctx.TX, t2)

// //     results, total, err := ctx.MusicService.SearchTracks(context.Background(), uniqueArtist1, 10, 0)

// //     assert.NoError(t, err)
// //     require.GreaterOrEqual(t, total, 1)
// //     assert.Equal(t, "Lose Yourself", results[0].Title)
// //     assert.Equal(t, uniqueArtist1, results[0].Artists)

// //     // Search by second unique artist
// //     results2, total2, err2 := ctx.MusicService.SearchTracks(context.Background(), uniqueArtist2, 10, 0)
// //     assert.NoError(t, err2)
// //     require.GreaterOrEqual(t, total2, 1)
// //     assert.Equal(t, "Shape of You", results2[0].Title)
// //     assert.Equal(t, uniqueArtist2, results2[0].Artists)
// // }

// // ==================== Like System Tests ====================

// // func TestLikeTrack_Success(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     ctx := setupMusicTests(t, db)
// //     defer ctx.Cleanup()

// //     user := createTestUser(newUUID(), "liker_"+newUUID()[:8], "like"+newUUID()+"@test.com")
// //     require.NoError(t, ctx.UserService.CreateUser(context.Background(), user))

// //     track := createTestTrackObject(newUUID(), "Liked Song", "Artist")
// //     seedTrack(t, ctx.TX, track)

// //     err := ctx.MusicService.LikeTrack(context.Background(), user.ID, track.ID)
// //     assert.NoError(t, err)

// //     liked, err := ctx.MusicService.IsTrackLiked(context.Background(), user.ID, track.ID)
// //     assert.NoError(t, err)
// //     assert.True(t, liked)
// // }

// func TestUnlikeTrack(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     ctx := setupMusicTests(t, db)
//     defer ctx.Cleanup()

//     user := createTestUser(newUUID(), "unliker_"+newUUID()[:8], "unlike"+newUUID()+"@test.com")
//     require.NoError(t, ctx.UserService.CreateUser(context.Background(), user))

//     track := createTestTrackObject(newUUID(), "Song", "Art")
//     seedTrack(t, ctx.TX, track)

//     require.NoError(t, ctx.MusicService.LikeTrack(context.Background(), user.ID, track.ID))

//     err := ctx.MusicService.UnlikeTrack(context.Background(), user.ID, track.ID)
//     assert.NoError(t, err)

//     liked, err := ctx.MusicService.IsTrackLiked(context.Background(), user.ID, track.ID)
//     assert.NoError(t, err)
//     assert.False(t, liked)
// }

// func TestGetLikedTracks(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     ctx := setupMusicTests(t, db)
//     defer ctx.Cleanup()

//     user := createTestUser(newUUID(), "listlikes_"+newUUID()[:8], "list"+newUUID()+"@test.com")
//     require.NoError(t, ctx.UserService.CreateUser(context.Background(), user))

//     for i := 0; i < 3; i++ {
//         track := createTestTrackObject(newUUID(), "S", "A")
//         seedTrack(t, ctx.TX, track)
//         require.NoError(t, ctx.MusicService.LikeTrack(context.Background(), user.ID, track.ID))
//     }

//     likes, total, err := ctx.MusicService.GetLikedTracks(context.Background(), user.ID, 10, 0)
//     assert.NoError(t, err)
//     assert.Equal(t, 3, total)
//     assert.Len(t, likes, 3)
// }

// // ==================== History Tests ====================

// func TestHistory_Flow(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     ctx := setupMusicTests(t, db)
//     defer ctx.Cleanup()

//     user := createTestUser(newUUID(), "historyuser_"+newUUID()[:8], "hist"+newUUID()+"@test.com")
//     require.NoError(t, ctx.UserService.CreateUser(context.Background(), user))

//     track := createTestTrackObject(newUUID(), "History Song", "Artist")
//     seedTrack(t, ctx.TX, track)

//     err := ctx.MusicService.AddTrackToHistory(context.Background(), user.ID, track.ID)
//     assert.NoError(t, err)

//     hist, total, err := ctx.MusicService.GetTrackHistory(context.Background(), user.ID, 10, 0)
//     assert.NoError(t, err)
//     assert.Equal(t, 1, total)
//     assert.Equal(t, track.ID, hist[0].TrackID)

//     err = ctx.MusicService.ClearTrackHistory(context.Background(), user.ID)
//     assert.NoError(t, err)

//     histAfter, totalAfter, err := ctx.MusicService.GetTrackHistory(context.Background(), user.ID, 10, 0)
//     assert.NoError(t, err)
//     assert.Equal(t, 0, totalAfter)
//     assert.Len(t, histAfter, 0)
// }