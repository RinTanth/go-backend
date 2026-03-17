package member

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/RinTanth/go-backend/app/member/access"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/wrapper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetMemberRequest struct {
	MemberID uuid.UUID `json:"memberId" binding:"required"`
}

type GetMemberResponse struct {
	MemberID    uuid.UUID               `json:"memberId"`
	Username    string                  `json:"username"`
	Email       string                  `json:"email"`
	HashedEmail string                  `json:"hashedEmail"`
	Status      access.MemberStatusType `json:"status"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
}

func (h *handler) GetInfo(c *gin.Context) {

	ctx := c.Request.Context()

	var req GetMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wrapper.Respond(c, wrapper.ResponseOption[GetMemberResponse]{
			HTTPStatus: http.StatusBadRequest,
			Code:       app.CodeBadRequest,
			Message:    app.MessageBadRequest,
			Err:        err,
		})
		return
	}

	memberInfo, err := h.memberStorage.GetMemberById(ctx, req.MemberID)
	if err != nil {
		slog.Error("fail to get member info", slog.String("err", err.Error()), slog.String("tag", "get member info"))
		wrapper.Respond(c, wrapper.ResponseOption[GetMemberResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	email, err := h.aesgcm.Decrypt(memberInfo.EncryptedEmail)
	if err != nil {
		slog.Error("fail to decrypt member email", slog.String("err", err.Error()), slog.String("tag", "get member info"))
		wrapper.Respond(c, wrapper.ResponseOption[GetMemberResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
		})
		return
	}

	wrapper.Respond(c, wrapper.ResponseOption[GetMemberResponse]{
		HTTPStatus: http.StatusOK,
		Code:       app.CodeSuccess,
		Message:    app.MessageSuccess,
		Data: &GetMemberResponse{
			MemberID:    memberInfo.GetID(),
			Username:    memberInfo.Username,
			Email:       email,
			HashedEmail: memberInfo.HashedEmail,
			Status:      memberInfo.Status,
			CreatedAt:   memberInfo.CreatedAt,
			UpdatedAt:   memberInfo.UpdatedAt,
		},
	})

}
