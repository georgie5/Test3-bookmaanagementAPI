package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

// The Set() method computes the hash of the password
func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

// Compare the client-provided plaintext password with saved-hashed version
func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// Check that a valid password is provided
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {

	v.Check(user.Username != "", "username", "must be provided")
	v.Check(len(user.Username) <= 200, "username", "must not be more than 200 bytes long")

	// validate email for user
	ValidateEmail(v, user.Email)

	// validate the plain text password
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	// check if we messed up in our codebase
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}

}

// Specify a custom duplicate email error message
var ErrDuplicateEmail = errors.New("duplicate email")

type UserModel struct {
	DB *sql.DB
}

// Insert a new user into the database
func (u UserModel) Insert(user *User) error {
	// the SQL query to be executed against the database table
	query := `
		INSERT INTO users (username, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`
	// the actual values to replace $1, $2, $3, and $4
	args := []any{user.Username, user.Email, user.Password.hash, user.Activated}
	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// if an email address already exists we will get a pq error message
	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// Get a user from the database based on their email provided
func (u UserModel) GetByEmail(email string) (*User, error) {
	// the SQL query to be executed against the database table
	query := `
		SELECT id, created_at, username, email, password_hash, activated, version
		FROM users
		WHERE email = $1
	`
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update a User. If the version number is different
// than what is was before the ran the query, it means
// someone did a previous edit or is doing an edit, so
// our query will fail and we would need to try again a bit later
func (u UserModel) Update(user *User) error {
	// the SQL query to be executed against the database table
	query := `
		 UPDATE users
		 SET username = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
		 WHERE id = $5 AND version = $6
		 RETURNING version
	 `
	// the actual values to replace $1, $2, $3, $4, $5, and $6
	args := []any{user.Username, user.Email, user.Password.hash, user.Activated, user.ID, user.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)

	// Check for errors during update
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Verify token to user. We need to hash the passed in token

func (u UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// We will do a join- I hope you still remember how to do a join
	query := `
	SELECT users.id, users.created_at, users.username,users.email, users.password_hash, users.activated, users.version
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1
	AND tokens.scope = $2 
	AND tokens.expiry > $3
	`

	args := []any{tokenHash[:], tokenScope, time.Now()}
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := u.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Username,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Return the matching user.
	return &user, nil
}

// /api/v1/users/{id}         # Get user profile
func (u UserModel) GetUser(id int64) (*User, error) {
	// the SQL query to be executed against the database table
	query := `
		SELECT id, created_at, username, email, activated
		FROM users
		WHERE id = $1
	`
	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a new User struct to hold the data returned by the query
	user := &User{}
	// Execute the query and scan the returned row into the User struct
	err := u.DB.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.CreatedAt, &user.Username, &user.Email, &user.Activated)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GET    /api/v1/users/{id}/lists   # Get user's reading lists
func (u UserModel) GetLists(id int64) ([]*ReadingList, error) {
	// the SQL query to be executed against the database table
	query := `
		SELECT id, name, description, status, created_by, version
		FROM reading_lists
		WHERE created_by = $1
	`
	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query and scan the returned rows into a slice of List structs
	rows, err := u.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a new slice to hold the List structs
	lists := []*ReadingList{}
	// Iterate over the rows returned by the query
	for rows.Next() {
		list := &ReadingList{}
		// Scan the current row into the List struct
		err := rows.Scan(&list.ID, &list.Name, &list.Description, &list.Status, &list.CreatedBy, &list.Version)
		if err != nil {
			return nil, err
		}
		// Append the List struct to the slice
		lists = append(lists, list)
	}

	// Check for errors during the iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return lists, nil
}

// GET    /api/v1/users/{id}/reviews # Get user's reviews
func (u UserModel) GetReviews(id int64) ([]*Review, error) {
	// the SQL query to be executed against the database table
	query := `
		SELECT id, review_date, book_id, rating, review
		FROM reviews
		WHERE user_id = $1
	`
	// Create a context with a 3-second timeout. No database
	// operation should take more than 3 seconds or we will quit it
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute the query and scan the returned rows into a slice of Review structs
	rows, err := u.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Create a new slice to hold the Review structs
	reviews := []*Review{}
	// Iterate over the rows returned by the query
	for rows.Next() {
		review := &Review{}
		// Scan the current row into the Review struct
		err := rows.Scan(&review.ID, &review.ReviewDate, &review.BookID, &review.Rating, &review.Review)
		if err != nil {
			return nil, err
		}
		// Append the Review struct to the slice
		reviews = append(reviews, review)
	}

	// Check for errors during the iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}
