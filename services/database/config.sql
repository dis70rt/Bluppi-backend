CREATE TABLE tracks (
    id INT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    album VARCHAR(255),
    duration INT,
    genre TEXT[],
    image_url VARCHAR(255),
    preview_url VARCHAR(255),
    video_id VARCHAR(255),
    listeners INT DEFAULT 0,
    play_count INT DEFAULT 0,
    popularity INT DEFAULT 0
)

