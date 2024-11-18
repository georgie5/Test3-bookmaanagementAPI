CREATE TABLE IF NOT EXISTS reviews (
    id bigserial PRIMARY KEY,
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    rating INT CHECK (rating >= 1 AND rating <= 5),
    review TEXT NOT NULL,
    review_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1
);