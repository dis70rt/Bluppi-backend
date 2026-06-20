# YouTube Music Migration Tasks

- [ ] **Phase 1: Cleanup & Branching**
  - [x] Create a new git branch: `git checkout -b feature/youtube-music-migration`
  - [ ] Delete the Solr implementation file: `rm internals/infrastructure/database/solr.go`
  - [ ] Remove Solr references from `internals/music/repository.go` (remove `SolrSearcher` interface and field).
  - [ ] Remove the Solr container from `docker-compose.yml`.

- [ ] **Phase 2: The Python Microservice (FastAPI)**
  - [ ] Initialize a new Python project (e.g., in a new `python-service/` folder alongside `bluppi-backend`).
  - [ ] Install dependencies: `pip install fastapi uvicorn ytmusicapi yt-dlp pydantic`.
  - [ ] Create `main.py` with FastAPI setup.
  - [ ] Implement `GET /search?q={query}` endpoint using `ytmusicapi.search()`. Map raw results to a clean Pydantic model (ID, Title, Artists, Thumbnails, Duration).
  - [ ] Implement `GET /stream/{youtube_id}` endpoint using `yt-dlp` to extract the direct `.m4a` audio URL.
  - [ ] Implement `GET /recommend/{youtube_id}` using `ytmusicapi.get_watch_playlist()` to get related tracks.
  - [ ] Dockerize the Python service and add it to `docker-compose.yml` under an internal network.

- [ ] **Phase 3: Go Backend - API Integration & Caching**
  - [ ] Create a new package/file in Go: `internals/infrastructure/youtube/client.go`.
  - [ ] Implement a Go HTTP client with `http.Transport` (connection pooling) to talk to the Python microservice.
  - [ ] Add Redis caching to the Go `youtube.Client`. Cache `/search` queries for 2 hours and stream URLs for 1 hour.
  - [ ] Update `internals/music/service.go` -> `SearchTracks()` to call your new Go `youtube.Client` instead of Solr. Map the JSON response to your gRPC `SearchResponse`.

- [ ] **Phase 4: The "Metadata Cache" (Upsert Pattern)**
  - [ ] In `internals/music/repository.go`, create an `UpsertTrack(ctx, track)` function that runs: `INSERT INTO tracks ... ON CONFLICT (track_id) DO NOTHING`.
  - [ ] In `internals/music/service.go` -> `PlayTrack` (or wherever playback/history is logged), launch a goroutine: `go func() { s.repo.UpsertTrack(bgCtx, track); s.repo.AddTrackToHistory(bgCtx, userID, track.ID) }()`.
  - [ ] Do the same goroutine upsert for `LikeTrack`.
  - [ ] Update playback endpoint to return the direct YouTube stream URL (fetched from the Go `youtube.Client` -> Python service) as a 302 Redirect to the client.

- [ ] **Phase 5: Hybrid Recommendation Engine**
  - [ ] Modify `WeeklyDiscoverTracks` in `internals/music/service.go`.
  - [ ] Step 5a: Fetch social tracks from Memgraph (`s.graphRepo.GetWeeklyDiscover`).
  - [ ] Step 5b: Fetch user's top 2 recent tracks from PostgreSQL history.
  - [ ] Step 5c: Pass those 2 tracks to the Python service `/recommend` to get algorithmic tracks.
  - [ ] Step 5d: Merge, interleave, and boost overlapping tracks in Go before returning them to the client.

- [ ] **Phase 6: Infrastructure & Queues (Optional but Recommended)**
  - [ ] Install `github.com/hibiken/asynq` in the Go backend.
  - [ ] Set up Asynq workers in `cmd/api/main.go` using your existing Redis connection.
  - [ ] Refactor Push Notifications (`notifications/consumer.go`) to push an `asynq.Task` instead of calling Firebase directly in the event loop, ensuring reliable retries.
