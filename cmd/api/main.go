package main

import (
	"context"
	"database/sql"
	"flag"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/georgie5/Test3-bookclubapi/internal/data"
	"github.com/georgie5/Test3-bookclubapi/internal/mailer"
	_ "github.com/lib/pq" // PostgreSQL driver
)

const appVersion = "1.0.0"

type serverConfig struct {
	port        int
	environment string
	db          struct {
		dsn string
	}

	limiter struct {
		rps     float64 // requests per second
		burst   int     // initial requests possible
		enabled bool    // enable or disable rate limiter
	}

	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

type applicationDependencies struct {
	config               serverConfig
	logger               *slog.Logger
	bookModel            *data.BookModel
	AuthorModel          *data.AuthorModel
	BookAuthorModel      *data.BookAuthorModel
	readingListModel     *data.ReadingListModel
	readingListBookModel *data.ReadingListBookModel
	reviewModel          *data.ReviewModel
	userModel            *data.UserModel
	mailer               mailer.Mailer
	wg                   sync.WaitGroup // need this later for background jobs
	tokenModel           data.TokenModel
}

func main() {

	var settings serverConfig

	flag.IntVar(&settings.port, "port", 4000, "Server port")
	flag.StringVar(&settings.environment, "env", "development", "Environment(developmnet|staging|production)")

	flag.StringVar(&settings.db.dsn, "db-dsn", "postgres://bookclub:bookclub@localhost/bookclub?sslmode=disable", "PostgreSQL DSN")

	flag.Float64Var(&settings.limiter.rps, "limiter-rps", 2, "Rate Limiter maximum requests per second")

	flag.IntVar(&settings.limiter.burst, "limiter-burst", 5, "Rate Limiter maximum burst")

	flag.BoolVar(&settings.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&settings.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	// We have port 25, 465, 587, 2525. If 25 doesn't work choose another
	flag.IntVar(&settings.smtp.port, "smtp-port", 587, "SMTP port")
	// Use your Username value provided by Mailtrap
	flag.StringVar(&settings.smtp.username, "smtp-username", "5f68753fd3d8ce", "SMTP username")

	// Use your Password value provided by Mailtrap
	flag.StringVar(&settings.smtp.password, "smtp-password", "f8e72801757fc0", "SMTP password")

	flag.StringVar(&settings.smtp.sender, "smtp-sender", "Comments Community <no-reply@commentscommunity.georgie.net>", "SMTP sender")

	flag.Parse()

	// Initialize the logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Set up the database connection (optional for now since we’re not using it)
	db, err := openDB(settings)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("Database connection pool established")

	// Initialize application dependencies
	appInstance := &applicationDependencies{
		config:               settings,
		logger:               logger,
		bookModel:            &data.BookModel{DB: db},
		AuthorModel:          &data.AuthorModel{DB: db},
		BookAuthorModel:      &data.BookAuthorModel{DB: db},
		readingListModel:     &data.ReadingListModel{DB: db},
		readingListBookModel: &data.ReadingListBookModel{DB: db},
		reviewModel:          &data.ReviewModel{DB: db},
		userModel:            &data.UserModel{DB: db},
		mailer:               mailer.New(settings.smtp.host, settings.smtp.port, settings.smtp.username, settings.smtp.password, settings.smtp.sender),
		tokenModel:           data.TokenModel{DB: db},
	}

	err = appInstance.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

}

// openDB sets up a connection pool to the database
func openDB(settings serverConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", settings.db.dsn)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
