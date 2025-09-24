package main

import (
	"time"

	"github.com/critma/goblog/internal/auth"
	"github.com/critma/goblog/internal/env"
	"github.com/critma/goblog/internal/store/postgres"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

// const version = "1.0.0"

func main() {
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Fatal("Error loading .env file")
	}

	config := setConfig()

	db, err := postgres.NewConnection(config.db.addr, config.db.maxOpenConns, config.db.maxIdleConns, config.db.maxIdleTime)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	store := postgres.NewStorage(db)

	JWTAuthenticator := auth.NewJWTAuthenticator(
		config.auth.secret, config.auth.issuer, config.auth.issuer,
	)

	app := &application{
		config:        *config,
		store:         store,
		logger:        logger,
		authenticator: JWTAuthenticator,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}

func setConfig() *config {
	return &config{
		addr: env.GetNonEmptyString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetNonEmptyString("DB_ADDR", "postgres://admin:admin@localhost/blog?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetNonEmptyString("DB_MAX_IDLE_TIME", "15m"),
		},
		auth: authConfig{
			secret: env.GetNonEmptyString("AUTH_SECRET", "secret"),
			issuer: env.GetNonEmptyString("AUTH_ISSUER", "blog"),
			exp:    time.Hour * 24,
		},
	}
}
