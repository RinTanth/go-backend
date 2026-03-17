package auth

import (
	"github.com/RinTanth/go-backend/app/auth/access"
	"github.com/RinTanth/go-common/aesgcm"
	"github.com/RinTanth/go-common/hash"
	"github.com/RinTanth/go-common/token"
	"github.com/jackc/pgx/v5/pgxpool"
)

type HandlerConfig struct {
	Pg           *pgxpool.Pool
	GoogleClient access.GoogleClienter
	Hash         hash.HashManager
	Aesgcm       aesgcm.Aesgcm
	Token        token.JWTSigner
}

type handler struct {
	pg           *pgxpool.Pool
	googleClient access.GoogleClienter
	hash         hash.HashManager
	aesgcm       aesgcm.Aesgcm
	token        token.JWTSigner
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{
		pg:           cfg.Pg,
		googleClient: cfg.GoogleClient,
		hash:         cfg.Hash,
		aesgcm:       cfg.Aesgcm,
		token:        cfg.Token,
	}
}
