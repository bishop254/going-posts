package main

import (
	"net/http"
	"time"

	"github.com/bishop254/bursary/internal/auth"
	"github.com/bishop254/bursary/internal/mailer"
	"github.com/bishop254/bursary/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type application struct {
	config        config
	store         store.Storage
	logger        *zap.SugaredLogger
	mailer        mailer.Client
	authenticator auth.Authenticator
	minio         *minio.Client
}

type config struct {
	addr  string
	db    dbConfig
	env   string
	mail  mailConfig
	auth  authConfig
	minio minioConfig
}

type authConfig struct {
	basicAuth basicAuthConfig
	jwtAuth   jwtConfig
}

type jwtConfig struct {
	secret string
	exp    time.Duration
	iss    string
	aud    string
}

type basicAuthConfig struct {
	username string
	password string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type mailConfig struct {
	tokenExp  time.Duration
	fromEmail string
	sendGrid  sendGridConfig
}

type sendGridConfig struct {
	apiKey string
}

type minioConfig struct {
	endpoint        string
	accessKeyID     string
	secretAccessKey string
	bucketName      string
	useSSL          bool
}

func (a *application) mount() http.Handler {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	r.Use(a.CorsMiddleware)

	r.Use(middleware.Timeout(time.Minute))

	r.Route("/v1", func(r chi.Router) {
		r.With(a.BasicAuthMiddleware()).Get("/health", a.healthCheckHandler)

		r.Route("/posts", func(r chi.Router) {
			r.Use(a.JWTAuthMiddleware())
			r.Post("/", a.createPostHandler)
			r.Post("/comment", a.createPostCommentHandler)

			r.Route("/{postId}", func(r chi.Router) {
				r.Use(a.postContextMiddleware)
				r.Get("/", a.getOnePostHandler)
				r.Delete("/", a.AuthorizationCheck("admin", a.deletePostHandler))
				r.Patch("/", a.AuthorizationCheck("moderator", a.updatePostHandler))
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Route("/activate/{token}", func(r chi.Router) {
				r.Get("/", a.activateUserHandler)
			})

			r.Route("/{userId}", func(r chi.Router) {
				r.Use(a.JWTAuthMiddleware())
				r.Use(a.userContextMiddleware)
				r.Get("/", a.getOneUserHandler)
				r.Put("/follow", a.followUserHandler)
				r.Put("/unfollow", a.unfollowUserHandler)
			})

			r.Group(func(r chi.Router) {
				r.Use(a.JWTAuthMiddleware())
				r.Get("/feed", a.getUserFeedHandler)
			})
		})

		r.Route("/auth", func(r chi.Router) {
			r.Post("/user", a.createUserHandler)
			r.With(a.BasicAuthMiddleware()).Post("/login", a.loginUserHandler)
		})
	})

	r.Route("/v8", func(r chi.Router) {
		//Admin
		r.Route("/auth-adm", func(r chi.Router) {
			r.With(a.BasicAuthMiddleware()).Post("/login", a.loginAdminHandler)
			r.Post("/register", a.registerAdminHandler)
		})

		r.Route("/admins", func(r chi.Router) {
			r.Route("/activate/{token}", func(r chi.Router) {
				r.Get("/", a.activateAdminHandler)
			})

			r.Route("/bursary", func(r chi.Router) {
				r.Use(a.JWTAuthMiddleware())
				r.Get("/", a.getBursariesHandler)
				r.Post("/", a.createBursaryHandler)
				r.Put("/", a.updateBursaryHandler)
			})
		})

		//Students
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", a.registerStudentHandler)
			r.With(a.BasicAuthMiddleware()).Post("/login", a.loginStudentHandler)
		})

		r.Route("/students", func(r chi.Router) {
			r.Route("/activate/{token}", func(r chi.Router) {
				r.Get("/", a.activateStudentHandler)
			})

			r.Route("/bursary", func(r chi.Router) {
				r.Use(a.JWTStudentAuthMiddleware())
				r.Get("/", a.getBursariesHandler)
				r.Get("/{bursaryID}", a.getBursaryByIDHandler)
			})

			r.Route("/{studentId}", func(r chi.Router) {
				r.Use(a.JWTStudentAuthMiddleware())
				r.Get("/personal", a.getStudentPersonalHandler)
				r.Post("/personal", a.createStudentPersonalHandler)
				r.Put("/personal", a.updateStudentPersonalHandler)

				r.Get("/institution", a.getStudentInstitutionHandler)
				r.Post("/institution", a.createStudentInstitutionHandler)
				r.Put("/institution", a.updateStudentInstitutionHandler)

				r.Get("/sponsor", a.getStudentSponsorHandler)
				r.Post("/sponsor", a.createStudentSponsorHandler)
				r.Put("/sponsor", a.updateStudentSponsorHandler)

				r.Get("/guardians", a.getStudentGuardiansHandler)
				r.Post("/guardians", a.createStudentGuardiansHandler)
				r.Put("/guardians", a.updateStudentGuardiansHandler)
				r.Post("/guardian/delete", a.deleteStudentGuardiansHandler)

				r.Get("/emergency", a.getStudentEmergencyHandler)
				r.Post("/emergency", a.createStudentEmergencyHandler)
				r.Put("/emergency", a.updateStudentEmergencyHandler)

				//TODO : finish doc upload, make it a for loop and dynamic based on the multiple files being uploaded
				r.Post("/documents", a.uploadDocumentsHandler)

				r.Post("/application", a.createStudentApplicationHandler)
				r.Put("/application", a.withdrawStudentApplicationHandler)
				r.Get("/applications", a.getStudentApplicationsHandler)
			})
		})
	})

	return r
}

func (a *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         a.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Minute,
	}

	a.logger.Infow("Server started on port ", "addr", a.config.addr)

	return srv.ListenAndServe()
}
