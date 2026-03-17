package auth

import (
	"log/slog"
	"net/http"

	"github.com/RinTanth/go-backend/app/auth/access"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/token"
	"github.com/RinTanth/go-common/wrapper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type IssueTokenRequest struct {
	MemberID       uuid.UUID                         `json:"memberId" binding:"required"`
	OrganizationID uuid.UUID                         `json:"organizationId" binding:"required"`
	MemberRole     access.OrganizationMemberRoleType `json:"memberRole" binding:"required"`
}

type IssueTokenResponse struct {
	AccessToken string `json:"accessToken"`
}

func (h *handler) IssueToken(c *gin.Context) {
	var req IssueTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wrapper.Respond(c, wrapper.ResponseOption[IssueTokenResponse]{
			HTTPStatus: http.StatusBadRequest,
			Code:       app.CodeBadRequest,
			Message:    app.MessageBadRequest,
			Err:        err,
		})
		return
	}

	claims := token.Claims{
		Sub:            req.MemberID.String(),
		Jti:            uuid.New().String(),
		OrganizationID: req.OrganizationID.String(),
		MemberRole:     string(req.MemberRole),
	}

	accessToken, err := h.token.SignES256(claims)
	if err != nil {
		slog.Error("failed to sign token", slog.String("err", err.Error()), slog.String("tag", "issue token"))
		wrapper.Respond(c, wrapper.ResponseOption[IssueTokenResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	wrapper.Respond(c, wrapper.ResponseOption[IssueTokenResponse]{
		HTTPStatus: http.StatusOK,
		Code:       app.CodeSuccess,
		Message:    app.MessageSuccess,
		Data: &IssueTokenResponse{
			AccessToken: accessToken,
		},
	})
}
