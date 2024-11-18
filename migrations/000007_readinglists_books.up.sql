CREATE TABLE IF NOT EXISTS reading_lists_books (
    reading_list_id INT REFERENCES reading_lists(id) ON DELETE CASCADE,
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    PRIMARY KEY (reading_list_id, book_id)
);

CREATE INDEX IF NOT EXISTS idx_reading_lists_books_reading_list_id_book_id ON reading_lists_books(reading_list_id, book_id);
