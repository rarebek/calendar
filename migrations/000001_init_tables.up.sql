CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(256) UNIQUE,
    username VARCHAR(64) UNIQUE,
    password TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('Asia/Tashkent', CURRENT_TIMESTAMP) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(256),
    description TEXT,
    event_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('Asia/Tashkent', CURRENT_TIMESTAMP) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);


CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY,
    file_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT timezone('Asia/Tashkent', CURRENT_TIMESTAMP) NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS events_files (
    event_id UUID REFERENCES events(id) ON DELETE CASCADE,
    file_id UUID REFERENCES files(id) ON DELETE CASCADE
);
