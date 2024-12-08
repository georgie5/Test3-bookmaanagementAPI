# API Endpoints

This document provides `curl` commands for interacting with the API endpoints.

## Healthcheck

### Check Server Health

```sh
curl -X GET http://localhost:4000/v1/healthcheck
```

### Book routes

#### Search Books

```sh
curl -X GET http://localhost:4000/api/v1/books/search -H "Authorization: Bearer YOUR_TOKEN"
```

#### List Books

```sh
curl -X GET http://localhost:4000/v1/books -H "Authorization: Bearer YOUR_TOKEN"
```

#### Create Book

```sh
curl -X POST http://localhost:4000/v1/books -d '{
  "title": "Good Omens",
  "authors": ["Neil Gaiman", "Terry Pratchett"],
  "isbn": "9780060853983",
  "publication_date": "1990-05-01",
  "genre": "Fantasy",
  "description": "A comedic novel about the birth of the son of Satan and the coming of the End Times."
}' -H "Content-Type: application/json" -H "Authorization: Bearer YOUR_AUTH_TOKEN"
```

#### Get Book

```sh
curl -X GET http://localhost:4000/v1/books/:book_id -H "Authorization: Bearer YOUR_TOKEN"
```

#### Update Book

```sh
curl -X PUT http://localhost:4000/v1/books/:book_id -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{
    "title": "Updated Book Title",
    "author": "Updated Author Name",
    "isbn": "9780060853983",
    "publication_date": "1990-05-01",
    "genre": "Updated Genre",
    "description": "Updated Description."
}'
```

#### Delete Book

```sh
curl -X DELETE http://localhost:4000/v1/books/:book_id -H "Authorization: Bearer YOUR_TOKEN"
```

### Reading List routes

#### List Reading Lists

```sh
curl -X GET http://localhost:4000/api/v1/lists -H "Authorization: Bearer YOUR_TOKEN"
```

#### Get Reading List

```sh
curl -X GET http://localhost:4000/api/v1/lists/:list_id -H "Authorization: Bearer YOUR_TOKEN"
```

#### Create Reading List

```sh
curl -X POST http://localhost:4000/api/v1/lists -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{
    "name": "Reading List Name",
    "description": "Description of the reading list"
}'
```

#### Update Reading List

```sh
curl -X PUT http://localhost:4000/api/v1/lists/:list_id -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{
    "name": "Updated Reading List Name",
    "description": "Updated description of the reading list"
}'
```

#### Delete Reading List

```sh
curl -X DELETE http://localhost:4000/api/v1/lists/:list_id -H "Authorization: Bearer YOUR_TOKEN"
```

#### Add Book to Reading List

```sh
curl -X POST http://localhost:4000/api/v1/lists/:list_id/books -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{
    "book_id": "book_id"
}'
```

#### Remove Book from Reading List

```sh
curl -X DELETE http://localhost:4000/api/v1/lists/:list_id/books -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{
    "book_id": "book_id"
}'
```

### Review routes

#### Get Reviews for Book

```sh
curl -X GET http://localhost:4000/v1/books/:book_id/reviews -H "Authorization: Bearer YOUR_TOKEN"
```

#### Add Review

```sh
curl -X POST http://localhost:4000/v1/books/:book_id/reviews -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{
    "rating": 5,
    "comment": "Great book!"
}'
```

#### Update Review

```sh
curl -X PUT http://localhost:4000/v1/reviews/:review_id -H "Authorization: Bearer YOUR_TOKEN" -H "Content-Type: application/json" -d '{
    "rating": 4,
    "comment": "Updated review comment"
}'
```

#### Delete Review

```sh
curl -X DELETE http://localhost:4000/v1/reviews/:review_id -H "Authorization: Bearer YOUR_TOKEN"
```

### User routes

#### Register User

```sh
curl -X POST http://localhost:4000/v1/users -H "Content-Type: application/json" -d '{
    "username": "newuser",
    "email": "newuser@example.com",
    "password": "password123"
}'
```

#### Get User Profile

```sh
curl -X GET http://localhost:4000/v1/users/:user_id -H "Authorization: Bearer YOUR_TOKEN"
```

#### Get User Reading Lists

```sh
curl -X GET http://localhost:4000/v1/users/:user_id/lists -H "Authorization: Bearer YOUR_TOKEN"
```

#### Get User Reviews

```sh
curl -X GET http://localhost:4000/v1/users/:user_id/reviews -H "Authorization: Bearer YOUR_TOKEN"
```

#### Activate User

```sh
curl -X PUT http://localhost:4000/v1/users/activated -H "Content-Type: application/json" -d '{
   "token": "activation_code"
}'
```

#### Create Authentication Token

```sh
curl -X POST http://localhost:4000/v1/tokens/authentication -H "Content-Type: application/json" -d '{
    "email": "username",
    "password": "password"
}'
```

#### Create Password Reset Token

```sh
curl -X POST http://localhost:4000/v1/tokens/password-reset -H "Content-Type: application/json" -d '{
    "email": "user@example.com"
}'
```

#### Update User Password

```sh
curl -X PUT http://localhost:4000/v1/users/password -H "Content-Type: application/json" -d '{
  "password": "newpassword",
  "token": "YOUR_PASSWORD_RESET_TOKEN"
}'
```
