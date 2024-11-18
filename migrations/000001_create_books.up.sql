CREATE TABLE books (
    id bigserial PRIMARY KEY,
    title TEXT NOT NULL,
    isbn TEXT UNIQUE,
    publication_date DATE,
    genre TEXT NOT NULL,
    description TEXT NOT NULL,
    average_rating REAL DEFAULT 0,
    version integer NOT NULL DEFAULT 1
);