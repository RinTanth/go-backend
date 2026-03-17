package member

import (
	"github.com/RinTanth/go-backend/app/member/access"
	"github.com/RinTanth/go-common/aesgcm"
	"github.com/RinTanth/go-common/hash"
)

type HandlerConfig struct {
	MemberStorage access.MemberStorage
	Aesgcm        aesgcm.Aesgcm
	Hash          hash.HashManager
}

type handler struct {
	memberStorage access.MemberStorage
	aesgcm        aesgcm.Aesgcm
	hash          hash.HashManager
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{
		memberStorage: cfg.MemberStorage,
		aesgcm:        cfg.Aesgcm,
		hash:          cfg.Hash,
	}
}
