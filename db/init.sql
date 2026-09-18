CREATE TABLE IF NOT EXISTS notes (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title VARCHAR(200) NOT NULL CHECK (btrim(title) <> ''),
    content TEXT NOT NULL CHECK (btrim(content) <> ''),
    author VARCHAR(100) NOT NULL CHECK (btrim(author) <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);