package organization

import (
	"github.com/RinTanth/go-backend/app/organization/access"
)

type HandlerConfig struct {
	OrganizationStorage access.OrganizationStorage
}

type handler struct {
	organizationStorage access.OrganizationStorage
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{
		organizationStorage: cfg.OrganizationStorage,
	}
}
