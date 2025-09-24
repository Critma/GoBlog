package main

import (
	"time"

	"github.com/critma/goblog/internal/auth"
	"github.com/critma/goblog/internal/store"
	"go.uber.org/zap"
)

type application struct {
	config        config
	logger        *zap.SugaredLogger
	store         store.Storage
	authenticator auth.Authenticator
}

type config struct {
	addr string
	db   dbConfig
	auth authConfig
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type authConfig struct {
	secret string
	issuer string
	exp    time.Duration
}
