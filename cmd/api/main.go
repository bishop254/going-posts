package main

import (
	"time"

	"github.com/bishop254/bursary/internal/auth"
	"github.com/bishop254/bursary/internal/db"
	"github.com/bishop254/bursary/internal/mailer"
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
			tokenExp:  time.Hour * 24 * 3,
			fromEmail: goDotEnvVariable("FROM_EMAIL"),
			sendGrid: sendGridConfig{
				apiKey: goDotEnvVariable("SENDGRID_API_KEY"),
			},
		},
		auth: authConfig{
			basicAuth: basicAuthConfig{
				username: goDotEnvVariable("BASIC_USERNAME"),
				password: goDotEnvVariable("BASIC_PASSWORD"),
			},
			jwtAuth: jwtConfig{
				secret: goDotEnvVariable("JWT_SECRET"),
				exp:    time.Hour * 1,
			},
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

	mailer := mailer.NewSendGrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.jwtAuth.secret, "kcg", "kcg")

	store := store.NewStorage(db)
	app := &application{
		config:        cfg,
		store:         store,
		logger:        logger,
		mailer:        mailer,
		authenticator: &jwtAuthenticator,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))
}
