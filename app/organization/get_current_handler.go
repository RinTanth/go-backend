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

type GetCurrentRequest struct {
	MemberID uuid.UUID `json:"memberId" binding:"required"`
}

type GetCurrentResponse struct {
	OrganizationID uuid.UUID                         `json:"organizationId"`
	Name           string                            `json:"name"`
	Status         access.OrganizationStatusType     `json:"status"`
	Role           access.OrganizationMemberRoleType `json:"role"`
	JoinedAt       *time.Time                        `json:"joinedAt,omitempty"`
	CreatedAt      time.Time                         `json:"createdAt"`
	UpdatedAt      time.Time                         `json:"updatedAt"`
}

func (h *handler) GetCurrent(c *gin.Context) {
	ctx := c.Request.Context()

	var req GetCurrentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wrapper.Respond(c, wrapper.ResponseOption[GetCurrentResponse]{
			HTTPStatus: http.StatusBadRequest,
			Code:       app.CodeBadRequest,
			Message:    app.MessageBadRequest,
			Err:        err,
		})
		return
	}

	orgMember, err := h.organizationStorage.GetMemberOrganization(ctx, req.MemberID)
	if err != nil {
		slog.Error("fail to get member organization", slog.String("err", err.Error()), slog.String("tag", "get current organization"))
		wrapper.Respond(c, wrapper.ResponseOption[GetCurrentResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	org, err := h.organizationStorage.GetOrganizationByID(ctx, orgMember.GetOrganizationID())
	if err != nil {
		slog.Error("fail to get organization", slog.String("err", err.Error()), slog.String("tag", "get current organization"))
		wrapper.Respond(c, wrapper.ResponseOption[GetCurrentResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
		})
		return
	}

	wrapper.Respond(c, wrapper.ResponseOption[GetCurrentResponse]{
		HTTPStatus: http.StatusOK,
		Code:       app.CodeSuccess,
		Message:    app.MessageSuccess,
		Data: &GetCurrentResponse{
			OrganizationID: org.GetOrganizationID(),
			Name:           org.Name,
			Status:         org.Status,
			Role:           orgMember.Role,
			JoinedAt:       orgMember.JoinedAt,
			CreatedAt:      org.CreatedAt,
			UpdatedAt:      org.UpdatedAt,
		},
	})
}
