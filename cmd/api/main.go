package main

import (
	"time"

	"github.com/bishop254/bursary/internal/db"
	"github.com/bishop254/bursary/internal/store"
	"go.uber.org/zap"
)

const version = "0.0.1"

func main() {

	cfg := config{
		addr: goDotEnvVariable("ADDR"),
		db: dbConfig{
			addr:         goDotEnvVariable("DB_ADDR"),
			maxOpenConns: 30,
			maxIdleConns: 30,
			maxIdleTime:  goDotEnvVariable("DB_MAX_IDLE_TIME"),
		},
		env: goDotEnvVariable("ENV_MODE"),
		mail: mailConfig{
			tokenExp: time.Hour * 24 * 3,
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("DB connection established")

	store := store.NewStorage(db)
	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
