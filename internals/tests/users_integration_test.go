package tests

// import (
// 	"context"
// 	"fmt"
// 	"testing"
// 	"time"

// 	"github.com/dis70rt/bluppi-backend/internals/users"
// 	"github.com/jmoiron/sqlx"
// 	_ "github.com/lib/pq"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )



// func setupTestService(t *testing.T, db *sqlx.DB) (*users.Service, func()) {
//     t.Helper()

//     tx, err := db.Beginx()
//     require.NoError(t, err)

//     bus := &noOpPublisher{}
//     repo := users.NewRepositoryWithTx(tx)
//     service := users.NewService(repo, bus)

//     cleanup := func() {
//         tx.Rollback()
//     }

//     return service, cleanup
// }

// func createTestUser(id, username, email string) *users.User {
//     now := time.Now()
//     gender := "male"
//     return &users.User{
//         ID:             id,
//         Email:          email,
//         Username:       username,
//         Name:           "Test User",
//         Bio:            StringPtr("Test bio"),
//         Country:        StringPtr("USA"),
//         Phone:          StringPtr("+1234567890"),
//         ProfilePic:     nil,
//         FavoriteGenres: []string{"Rock", "Jazz"},
//         CreatedAt:      now,
//         DateOfBirth:    now,
//         Gender:         gender,
//     }
// }

// // ==================== CreateUser Tests ====================

// func TestCreateUser_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_1", "testuser1", "test1@example.com")

//     err := service.CreateUser(ctx, user)

//     assert.NoError(t, err)
// }

// func TestCreateUser_DuplicateEmail(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user1 := createTestUser("test_user_1", "testuser1", "duplicate@example.com")
//     err := service.CreateUser(ctx, user1)
//     require.NoError(t, err)

//     user2 := createTestUser("test_user_2", "testuser2", "duplicate@example.com")
//     err = service.CreateUser(ctx, user2)

//     assert.ErrorIs(t, err, users.ErrUserAlreadyExists)
// }

// func TestCreateUser_InvalidInput_NilUser(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.CreateUser(ctx, nil)

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// func TestCreateUser_InvalidInput_EmptyFields(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     testCases := []struct {
//         name string
//         user *users.User
//     }{
//         {"empty ID", &users.User{ID: "", Username: "user", Email: "e@e.com", Name: "N"}},
//         {"empty username", &users.User{ID: "id", Username: "", Email: "e@e.com", Name: "N"}},
//         {"empty email", &users.User{ID: "id", Username: "user", Email: "", Name: "N"}},
//         {"empty name", &users.User{ID: "id", Username: "user", Email: "e@e.com", Name: ""}},
//     }

//     for _, tc := range testCases {
//         t.Run(tc.name, func(t *testing.T) {
//             err := service.CreateUser(ctx, tc.user)
//             assert.ErrorIs(t, err, users.ErrInvalidInput)
//         })
//     }
// }

// // ==================== GetUserByID Tests ====================

// func TestGetUserByID_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_get", "getuser", "get@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     result, err := service.GetUserByID(ctx, "test_user_get")

//     assert.NoError(t, err)
//     assert.Equal(t, user.ID, result.ID)
//     assert.Equal(t, user.Username, result.Username)
//     assert.Equal(t, user.Email, result.Email)
// }

// func TestGetUserByID_NotFound(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.GetUserByID(ctx, "nonexistent_user")

//     assert.ErrorIs(t, err, users.ErrUserNotFound)
// }

// func TestGetUserByID_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.GetUserByID(ctx, "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== GetUserByUsername Tests ====================

// func TestGetUserByUsername_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_uname", "uniqueusername", "uname@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     result, err := service.GetUserByUsername(ctx, "uniqueusername")

//     assert.NoError(t, err)
//     assert.Equal(t, user.Username, result.Username)
// }

// func TestGetUserByUsername_NotFound(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.GetUserByUsername(ctx, "nonexistent_username")

//     assert.ErrorIs(t, err, users.ErrUserNotFound)
// }

// func TestGetUserByUsername_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.GetUserByUsername(ctx, "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== UpdateUser Tests ====================

// func TestUpdateUser_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_update", "updateuser", "update@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     updates := map[string]any{
//         "name":    "Updated Name",
//         "bio":     "Updated bio",
//         "country": "Canada",
//     }

//     err := service.UpdateUser(ctx, "test_user_update", updates)

//     assert.NoError(t, err)

//     updated, err := service.GetUserByID(ctx, "test_user_update")
//     assert.NoError(t, err)
//     assert.Equal(t, "Updated Name", updated.Name)
// }

// func TestUpdateUser_PartialUpdate(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_partial", "partialuser", "partial@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     updates := map[string]any{
//         "name": "Only Name Updated",
//     }

//     err := service.UpdateUser(ctx, "test_user_partial", updates)
//     assert.NoError(t, err)

//     updated, err := service.GetUserByID(ctx, "test_user_partial")
//     assert.NoError(t, err)
//     assert.Equal(t, "Only Name Updated", updated.Name)
//     assert.Equal(t, *user.Bio, *updated.Bio)
// }

// func TestUpdateUser_NotFound(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.UpdateUser(ctx, "nonexistent", map[string]any{"name": "New"})

//     assert.ErrorIs(t, err, users.ErrUserNotFound)
// }

// func TestUpdateUser_InvalidInput_EmptyID(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.UpdateUser(ctx, "", map[string]any{"name": "New"})

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// func TestUpdateUser_InvalidInput_EmptyUpdates(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.UpdateUser(ctx, "someid", map[string]any{})

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== DeleteUser Tests ====================

// func TestDeleteUser_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_delete", "deleteuser", "delete@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     err := service.DeleteUser(ctx, "test_user_delete")

//     assert.NoError(t, err)

//     _, err = service.GetUserByID(ctx, "test_user_delete")
//     assert.ErrorIs(t, err, users.ErrUserNotFound)
// }

// func TestDeleteUser_NotFound(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.DeleteUser(ctx, "nonexistent_delete")

//     assert.ErrorIs(t, err, users.ErrUserNotFound)
// }

// func TestDeleteUser_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.DeleteUser(ctx, "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== UsernameExists Tests ====================

// func TestUsernameExists_True(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_exists", "existsuser", "exists@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     exists, err := service.UsernameExists(ctx, "existsuser")

//     assert.NoError(t, err)
//     assert.True(t, exists)
// }

// func TestUsernameExists_False(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     exists, err := service.UsernameExists(ctx, "nonexistent_username")

//     assert.NoError(t, err)
//     assert.False(t, exists)
// }

// func TestUsernameExists_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.UsernameExists(ctx, "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== EmailExists Tests ====================

// func TestEmailExists_True(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()
//     user := createTestUser("test_user_email", "emailuser", "emailcheck@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     exists, err := service.EmailExists(ctx, "emailcheck@example.com")

//     assert.NoError(t, err)
//     assert.True(t, exists)
// }

// func TestEmailExists_False(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     exists, err := service.EmailExists(ctx, "nonexistent@example.com")

//     assert.NoError(t, err)
//     assert.False(t, exists)
// }

// func TestEmailExists_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.EmailExists(ctx, "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== SearchUsers Tests ====================

// func TestSearchUsers_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("search_user", "searchableuser", "search@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     results, total, err := service.SearchUsers(ctx, "searchable", 10, 0)

//     assert.NoError(t, err)
//     assert.GreaterOrEqual(t, len(results), 1)
//     assert.GreaterOrEqual(t, total, 1)
// }

// func TestSearchUsers_ByName(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := &users.User{
//         ID:             "name_search_user",
//         Email:          "namesearch@example.com",
//         Username:       "namesearchuser",
//         Name:           "UniqueNameForSearch",
//         FavoriteGenres: []string{},
//     }
//     require.NoError(t, service.CreateUser(ctx, user))

//     results, _, err := service.SearchUsers(ctx, "UniqueNameFor", 10, 0)

//     assert.NoError(t, err)
//     assert.GreaterOrEqual(t, len(results), 1)
// }

// func TestSearchUsers_NoResults(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     results, total, err := service.SearchUsers(ctx, "xyz_nonexistent_query_123", 10, 0)

//     assert.NoError(t, err)
//     assert.Len(t, results, 0)
//     assert.Equal(t, 0, total)
// }

// func TestSearchUsers_Pagination(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     for i := 0; i < 5; i++ {
//         user := createTestUser(
//             fmt.Sprintf("paginate_user_%d", i),
//             fmt.Sprintf("paginateuser%d", i),
//             fmt.Sprintf("paginate%d@example.com", i),
//         )
//         require.NoError(t, service.CreateUser(ctx, user))
//     }

//     results, total, err := service.SearchUsers(ctx, "paginateuser", 2, 0)

//     assert.NoError(t, err)
//     assert.Len(t, results, 2)
//     assert.Equal(t, 5, total)

//     results2, _, err := service.SearchUsers(ctx, "paginateuser", 2, 2)

//     assert.NoError(t, err)
//     assert.Len(t, results2, 2)
// }

// // ==================== Follow System Tests ====================
// // JUST SKIP THIS, TOO MUCH HASSLE
// // func TestFollow_Success(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     service, cleanup := setupTestService(t, db)
// //     defer cleanup()

// //     ctx := context.Background()

// //     user1 := createTestUser("follower_1", "follower1", "follower1@example.com")
// //     user2 := createTestUser("followee_1", "followee1", "followee1@example.com")
// //     require.NoError(t, service.CreateUser(ctx, user1))
// //     require.NoError(t, service.CreateUser(ctx, user2))

// //     err := service.Follow(ctx, "follower_1", "followee_1")

// //     assert.NoError(t, err)

// //     isFollowing, err := service.IsFollowing(ctx, "follower_1", "followee_1")
// //     assert.NoError(t, err)
// //     assert.True(t, isFollowing)
// // }

// func TestFollow_SelfFollow(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("self_follow", "selffollow", "self@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     err := service.Follow(ctx, "self_follow", "self_follow")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// func TestFollow_InvalidInput_EmptyFollowerID(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.Follow(ctx, "", "some_followee")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// func TestFollow_InvalidInput_EmptyFolloweeID(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.Follow(ctx, "some_follower", "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }
// // SKIP THIS TOO FUCK THIS>>>
// // func TestFollow_Duplicate(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     service, cleanup := setupTestService(t, db)
// //     defer cleanup()

// //     ctx := context.Background()

// //     user1 := createTestUser("dup_follower", "dupfollower", "dupfollower@example.com")
// //     user2 := createTestUser("dup_followee", "dupfollowee", "dupfollowee@example.com")
// //     require.NoError(t, service.CreateUser(ctx, user1))
// //     require.NoError(t, service.CreateUser(ctx, user2))

// //     require.NoError(t, service.Follow(ctx, "dup_follower", "dup_followee"))

// //     err := service.Follow(ctx, "dup_follower", "dup_followee")

// //     assert.NoError(t, err)
// // }
// // FUCK THIS TOO
// // func TestUnfollow_Success(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     service, cleanup := setupTestService(t, db)
// //     defer cleanup()

// //     ctx := context.Background()

// //     user1 := createTestUser("unfollower_1", "unfollower1", "unfollower1@example.com")
// //     user2 := createTestUser("unfollowee_1", "unfollowee1", "unfollowee1@example.com")
// //     require.NoError(t, service.CreateUser(ctx, user1))
// //     require.NoError(t, service.CreateUser(ctx, user2))
// //     require.NoError(t, service.Follow(ctx, "unfollower_1", "unfollowee_1"))

// //     err := service.Unfollow(ctx, "unfollower_1", "unfollowee_1")

// //     assert.NoError(t, err)

// //     isFollowing, err := service.IsFollowing(ctx, "unfollower_1", "unfollowee_1")
// //     assert.NoError(t, err)
// //     assert.False(t, isFollowing)
// // }

// func TestUnfollow_NotFollowing(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.Unfollow(ctx, "random_1", "random_2")

//     assert.ErrorIs(t, err, users.ErrNotFollowing)
// }

// func TestUnfollow_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.Unfollow(ctx, "", "some_followee")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }
// // FUCKKK
// // func TestIsFollowing_True(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     service, cleanup := setupTestService(t, db)
// //     defer cleanup()

// //     ctx := context.Background()

// //     user1 := createTestUser("isfollowing_1", "isfollowing1", "isfollowing1@example.com")
// //     user2 := createTestUser("isfollowing_2", "isfollowing2", "isfollowing2@example.com")
// //     require.NoError(t, service.CreateUser(ctx, user1))
// //     require.NoError(t, service.CreateUser(ctx, user2))
// //     require.NoError(t, service.Follow(ctx, "isfollowing_1", "isfollowing_2"))

// //     isFollowing, err := service.IsFollowing(ctx, "isfollowing_1", "isfollowing_2")

// //     assert.NoError(t, err)
// //     assert.True(t, isFollowing)
// // }

// func TestIsFollowing_False(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     isFollowing, err := service.IsFollowing(ctx, "notfollowing_1", "notfollowing_2")

//     assert.NoError(t, err)
//     assert.False(t, isFollowing)
// }

// func TestIsFollowing_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.IsFollowing(ctx, "", "some_followee")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // func TestGetFollowers_Success(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     service, cleanup := setupTestService(t, db)
// //     defer cleanup()

// //     ctx := context.Background()

// //     user1 := createTestUser("target_user", "targetuser", "target@example.com")
// //     user2 := createTestUser("follower_a", "followera", "followera@example.com")
// //     user3 := createTestUser("follower_b", "followerb", "followerb@example.com")
// //     require.NoError(t, service.CreateUser(ctx, user1))
// //     require.NoError(t, service.CreateUser(ctx, user2))
// //     require.NoError(t, service.CreateUser(ctx, user3))
// //     require.NoError(t, service.Follow(ctx, "follower_a", "target_user"))
// //     require.NoError(t, service.Follow(ctx, "follower_b", "target_user"))

// //     followers, total, err := service.GetFollowers(ctx, "target_user", 10, 0)

// //     assert.NoError(t, err)
// //     assert.Equal(t, 2, total)
// //     assert.Len(t, followers, 2)
// // }

// // func TestGetFollowers_Empty(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     service, cleanup := setupTestService(t, db)
// //     defer cleanup()

// //     ctx := context.Background()

// //     user := createTestUser("no_followers_user", "nofollowersuser", "nofollowers@example.com")
// //     require.NoError(t, service.CreateUser(ctx, user))

// //     followers, total, err := service.GetFollowers(ctx, "no_followers_user", 10, 0)

// //     assert.NoError(t, err)
// //     assert.Equal(t, 0, total)
// //     assert.Len(t, followers, 0)
// // }

// func TestGetFollowers_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, _, err := service.GetFollowers(ctx, "", 10, 0)

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // func TestGetFollowing_Success(t *testing.T) {
// //     db := GetTestDB(t)
// //     defer db.Close()

// //     service, cleanup := setupTestService(t, db)
// //     defer cleanup()

// //     ctx := context.Background()

// //     user1 := createTestUser("active_user", "activeuser", "active@example.com")
// //     user2 := createTestUser("followed_a", "followeda", "followeda@example.com")
// //     user3 := createTestUser("followed_b", "followedb", "followedb@example.com")
// //     require.NoError(t, service.CreateUser(ctx, user1))
// //     require.NoError(t, service.CreateUser(ctx, user2))
// //     require.NoError(t, service.CreateUser(ctx, user3))
// //     require.NoError(t, service.Follow(ctx, "active_user", "followed_a"))
// //     require.NoError(t, service.Follow(ctx, "active_user", "followed_b"))

// //     following, total, err := service.GetFollowing(ctx, "active_user", 10, 0)

// //     assert.NoError(t, err)
// //     assert.Equal(t, 2, total)
// //     assert.Len(t, following, 2)
// // }

// func TestGetFollowing_Empty(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("not_following_user", "notfollowinguser", "notfollowing@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     following, total, err := service.GetFollowing(ctx, "not_following_user", 10, 0)

//     assert.NoError(t, err)
//     assert.Equal(t, 0, total)
//     assert.Len(t, following, 0)
// }

// func TestGetFollowing_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, _, err := service.GetFollowing(ctx, "", 10, 0)

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== User Stats Tests ====================

// func TestGetUserStats_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("stats_user", "statsuser", "stats@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     stats, err := service.GetUserStats(ctx, "stats_user")

//     assert.NoError(t, err)
//     assert.NotNil(t, stats)
//     assert.Equal(t, 0, stats.LikedTracks)
//     assert.Equal(t, 0, stats.FollowersCount)
//     assert.Equal(t, 0, stats.FollowingCount)
// }

// func TestGetUserStats_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.GetUserStats(ctx, "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// // ==================== Recent Searches Tests ====================

// func TestAddRecentSearch_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("recent_user", "recentuser", "recent@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     err := service.AddRecentSearch(ctx, "recent_user", "test query")

//     assert.NoError(t, err)
// }

// func TestAddRecentSearch_InvalidInput_EmptyUserID(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.AddRecentSearch(ctx, "", "test query")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// func TestAddRecentSearch_InvalidInput_EmptyQuery(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     err := service.AddRecentSearch(ctx, "some_user", "")

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }

// func TestGetRecentSearches_Success(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("getsearch_user", "getsearchuser", "getsearch@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))
//     require.NoError(t, service.AddRecentSearch(ctx, "getsearch_user", "query 1"))
//     require.NoError(t, service.AddRecentSearch(ctx, "getsearch_user", "query 2"))
//     require.NoError(t, service.AddRecentSearch(ctx, "getsearch_user", "query 3"))

//     searches, err := service.GetRecentSearches(ctx, "getsearch_user", 10)

//     assert.NoError(t, err)
//     assert.Len(t, searches, 3)
// }

// func TestGetRecentSearches_Empty(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("nosearch_user", "nosearchuser", "nosearch@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     searches, err := service.GetRecentSearches(ctx, "nosearch_user", 10)

//     assert.NoError(t, err)
//     assert.Len(t, searches, 0)
// }

// func TestGetRecentSearches_Limit(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     user := createTestUser("limitsearch_user", "limitsearchuser", "limitsearch@example.com")
//     require.NoError(t, service.CreateUser(ctx, user))

//     for i := 0; i < 10; i++ {
//         require.NoError(t, service.AddRecentSearch(ctx, "limitsearch_user", fmt.Sprintf("query %d", i)))
//     }

//     searches, err := service.GetRecentSearches(ctx, "limitsearch_user", 5)

//     assert.NoError(t, err)
//     assert.Len(t, searches, 5)
// }

// func TestGetRecentSearches_InvalidInput(t *testing.T) {
//     db := GetTestDB(t)
//     defer db.Close()

//     service, cleanup := setupTestService(t, db)
//     defer cleanup()

//     ctx := context.Background()

//     _, err := service.GetRecentSearches(ctx, "", 10)

//     assert.ErrorIs(t, err, users.ErrInvalidInput)
// }