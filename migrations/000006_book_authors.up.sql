CREATE TABLE IF NOT EXISTS book_authors (
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    author_id INT REFERENCES authors(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);

CREATE INDEX IF NOT EXISTS idx_book_authors_book_id_author_id ON book_authors(book_id, author_id);
