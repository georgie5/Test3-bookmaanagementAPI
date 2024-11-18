-- Reading lists { id, name, description, created by, books, status (currently reading/completed)

CREATE TABLE reading_lists (
    id bigserial PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_by INT REFERENCES users(id) ON DELETE CASCADE,
    status TEXT DEFAULT 'currently reading' CHECK (status IN ('currently reading', 'completed')),
    version INTEGER NOT NULL DEFAULT 1
);
