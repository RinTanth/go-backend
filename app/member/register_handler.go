package member

import (
	"log/slog"
	"net/http"

	"github.com/RinTanth/go-backend/app/member/access"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/wrapper"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegisterMemberRequest struct {
	Email string `json:"email" binding:"required"`
	Name  string `json:"name" binding:"required"`
}

type RegisterMemberResponse struct {
	MemberID uuid.UUID `json:"memberId"`
}

func (h *handler) RegisterMember(c *gin.Context) {
	ctx := c.Request.Context()

	var req RegisterMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wrapper.Respond(c, wrapper.ResponseOption[RegisterMemberResponse]{
			HTTPStatus: http.StatusBadRequest,
			Code:       app.CodeBadRequest,
			Message:    app.MessageBadRequest,
			Err:        err,
		})
		return
	}

	// Hash email
	hashedEmail := h.hash.HashSha256EncodePepper(req.Email)

	// Generate IV for encryption
	iv, err := h.aesgcm.GenerateIV()
	if err != nil {
		slog.Error("fail to generate IV", slog.String("err", err.Error()), slog.String("tag", "register member"))
		wrapper.Respond(c, wrapper.ResponseOption[RegisterMemberResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	// Encrypt email
	encryptedEmail, err := h.aesgcm.Encrypt(req.Email, iv)
	if err != nil {
		slog.Error("fail to encrypt email", slog.String("err", err.Error()), slog.String("tag", "register member"))
		wrapper.Respond(c, wrapper.ResponseOption[RegisterMemberResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	member := access.Member{
		Username:       req.Name,
		EncryptedEmail: encryptedEmail,
		HashedEmail:    hashedEmail,
		Status:         access.MemberStatusActive,
	}

	createdMember, err := h.memberStorage.CreateMember(ctx, member)
	if err != nil {
		slog.Error("fail to create member", slog.String("err", err.Error()), slog.String("tag", "register member"))
		wrapper.Respond(c, wrapper.ResponseOption[RegisterMemberResponse]{
			HTTPStatus: http.StatusInternalServerError,
			Code:       app.CodeInternalError,
			Message:    app.MessageInternalError,
			Err:        err,
		})
		return
	}

	wrapper.Respond(c, wrapper.ResponseOption[RegisterMemberResponse]{
		HTTPStatus: http.StatusOK,
		Code:       app.CodeSuccess,
		Message:    app.MessageSuccess,
		Data: &RegisterMemberResponse{
			MemberID: createdMember.GetID(),
		},
	})
}
