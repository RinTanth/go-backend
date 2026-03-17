package organization

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/RinTanth/go-backend/app/organization/access"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/wrapper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterOrganizationRequest struct {
	Email    string    `json:"email" binding:"required"`
	Name     string    `json:"name" binding:"required"`
	MemberID uuid.UUID `json:"memberId" binding:"required"`
}

type RegisterOrganizationResponse struct {
	OrganizationID uuid.UUID                         `json:"organizationId"`
	MemberID       uuid.UUID                         `json:"memberId"`
	Role           access.OrganizationMemberRoleType `json:"role"`
}

func (h *handler) RegisterOrganization(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wrapper.Respond(c, wrapper.ResponseOption[RegisterOrganizationResponse]{
			HTTPStatus: http.StatusBadRequest,
			Code:       app.CodeBadRequest,
			Message:    app.MessageBadRequest,
			Err:        err,
		})
		return
	}

	// Create organization using provided email as name
	org := access.Organization{
		Name:   req.Email, // Use email as organization name
		Status: access.OrganizationStatusActive,
	}

	createdOrg, err := h.organizationStorage.CreateOrganization(ctx, org)
	if err != nil {
		slog.Error("fail to create organization", slog.String("err", err.Error()), slog.String("tag", "register organization"))
		wrapper.Respond(c, wrapper.ResponseOption[RegisterOrganizationResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	// Create organization-member relationship
	now := time.Now()
	orgMember := access.OrganizationMember{
		OrganizationID: createdOrg.OrganizationID,
		MemberID:       req.MemberID.String(),
		Role:           access.OrganizationMemberRoleAdmin,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.organizationStorage.CreateOrganizationMember(ctx, orgMember); err != nil {
		slog.Error("fail to create organization member", slog.String("err", err.Error()), slog.String("tag", "register organization"))
		wrapper.Respond(c, wrapper.ResponseOption[RegisterOrganizationResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	// Return organization ID, member ID, and role
	wrapper.Respond(c, wrapper.ResponseOption[RegisterOrganizationResponse]{
		HTTPStatus: http.StatusOK,
		Code:       app.CodeSuccess,
		Message:    app.MessageSuccess,
		Data: &RegisterOrganizationResponse{
			OrganizationID: createdOrg.GetOrganizationID(),
			MemberID:       req.MemberID,
			Role:           access.OrganizationMemberRoleAdmin,
		},
	})
}
