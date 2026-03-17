package router

import (
	"net/http"
	"time"

	"github.com/RinTanth/go-backend/app/auth"
	authaccess "github.com/RinTanth/go-backend/app/auth/access"
	"github.com/RinTanth/go-backend/config"
	"github.com/RinTanth/go-common/aesgcm"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/database"
	"github.com/RinTanth/go-common/hash"
	"github.com/RinTanth/go-common/health"
	"github.com/RinTanth/go-common/httpclient"
	"github.com/RinTanth/go-common/middleware"
	"github.com/RinTanth/go-common/token"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gin-gonic/gin"
)

// New constructs a gin.Engine with routes and middleware configured.
func New(cfg config.Config, version, commit string, timeoutDuration time.Duration) (*gin.Engine, func()) {
	r := gin.New()
	r.Use(gin.Recovery())

	if config.IsLocalEnv() {
		r.Use(gin.Logger())
	}

	// ctx := context.Background()

	r.GET("/liveness", health.Liveness(version, commit))
	r.GET("/metrics", health.Metrics())
	r.GET("/readiness", health.Readiness())

	r.Use(
		middleware.SecurityHeaders(),
		middleware.AccessControl(cfg.AccessControl.AllowOrigin, allowedHeaders(cfg.Header.RefIDHeaderKey)),
		app.TraceContextTraceIDMiddleware(""),
		app.RefIDMiddleware(cfg.Header.RefIDHeaderKey),
		app.AutoLoggingMiddleware,
		middleware.Timeout(timeoutDuration),
		middleware.AccessLog(),
	)

	httpClient := httpclient.NewHTTPClient(app.ForwardRefIDOption)
	hash := newHashManager(cfg)
	aesgcm := newAesgcm(cfg)
	tokenSigner := newTokenManager(cfg)
	db := newPostgresManager(cfg)

	registerAuthRoutes(r, httpClient, cfg, db, hash, aesgcm, tokenSigner)

	return r, func() {
		db.Close()
	}
}

func newPostgresManager(cfg config.Config) *pgxpool.Pool {
	return database.NewPostgresDB(database.PostgresConfig{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		DBName:   cfg.Postgres.Name,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
	})
}

func newTokenManager(cfg config.Config) token.JWTSigner {
	return token.MustNewJWTSigner(token.JWTSignerConfig{
		PrivateKey: cfg.JWT.PrivateKey,
		Alg:        string(token.ES256),
		Issuer:     cfg.JWT.Issuer,
		Audience:   cfg.JWT.Audience,
		Expire:     cfg.JWT.ExpDuration,
	})
}

func newHashManager(cfg config.Config) hash.HashManager {
	return hash.MustNewHashManager(hash.HashManagerCfgs{
		Pepper: cfg.Hash.Pepper,
	})
}

func newAesgcm(cfg config.Config) aesgcm.Aesgcm {
	return aesgcm.MustNewAesgcm(aesgcm.AesgcmCfgs{
		Key: cfg.Aesgcm.Key,
	})
}

func registerAuthRoutes(r *gin.Engine, httpClient *http.Client, cfg config.Config, pg *pgxpool.Pool, hash hash.HashManager, aesgcm aesgcm.Aesgcm, token token.JWTSigner) {
	googleClient := authaccess.NewGoogleClient(
		cfg.GoogleClient.VerifyTokenURL,
		cfg.GoogleClient.GetUserProfileURL,
		cfg.GoogleClient.RevokeTokenURL,
		httpClient,
	)

	authHandlerCfg := auth.HandlerConfig{
		Pg:           pg,
		GoogleClient: googleClient,
		Hash:         hash,
		Aesgcm:       aesgcm,
		Token:        token,
	}
	authHandler := auth.NewHandler(authHandlerCfg)

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/resolve-identity", authHandler.ResolveIdentify)
		authGroup.POST("/issue-token", authHandler.IssueToken)
	}
}

func allowedHeaders(refIDHeaderKey string) []string {
	return []string{
		"Content-Type",
		"Content-Length",
		"Accept-Encoding",
		"X-CSRF-Token",
		"Authorization",
		"accept",
		"origin",
		"Cache-Control",
		"X-Requested-With",
		refIDHeaderKey,
	}
}
