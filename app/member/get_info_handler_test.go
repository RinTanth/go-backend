package member_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RinTanth/go-backend/app/member"
	"github.com/RinTanth/go-backend/app/member/access"
	access_mocks "github.com/RinTanth/go-backend/app/member/access/mocks"
	aesgcm_mocks "github.com/RinTanth/go-common/aesgcm/mocks"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/wrapper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	r := require.New(t)

	memberID := uuid.New()
	now := time.Now()

	type mockArgs struct {
		memberStorage *access_mocks.MemberStorageMock
		aesgcm        *aesgcm_mocks.AesgcmMock
	}

	type args struct {
		ctx context.Context
		req member.GetMemberRequest
	}

	type want struct {
		err     bool
		code    app.Code
		Message app.Message
		data    wrapper.ResponseOption[member.GetMemberResponse]
	}

	tests := []struct {
		name    string
		prepare func(m mockArgs, args args)
		args    args
		want    want
	}{
		{
			name: "success, case valid request",
			prepare: func(m mockArgs, args args) {
				m.memberStorage.
					EXPECT().
					GetMemberById(args.ctx, args.req.MemberID).
					Return(access.Member{
						MemberID:       memberID.String(),
						Username:       "Test User",
						EncryptedEmail: "encrypted_email",
						HashedEmail:    "hashed_email",
						Status:         access.MemberStatusActive,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)

				m.aesgcm.
					EXPECT().
					Decrypt("encrypted_email").
					Return("test@example.com", nil)
			},
			args: args{
				req: member.GetMemberRequest{
					MemberID: memberID,
				},
			},
			want: want{
				err:     false,
				code:    app.CodeSuccess,
				Message: app.MessageSuccess,
				data: wrapper.ResponseOption[member.GetMemberResponse]{
					HTTPStatus: http.StatusOK,
					Code:       app.CodeSuccess,
					Message:    app.MessageSuccess,
					Data: &member.GetMemberResponse{
						MemberID:    memberID,
						Username:    "Test User",
						Email:       "test@example.com",
						HashedEmail: "hashed_email",
						Status:      access.MemberStatusActive,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				},
			},
		},
		{
			name: "fail, case invalid request body",
			prepare: func(m mockArgs, args args) {
				// no calls
			},
			args: args{
				req: member.GetMemberRequest{}, // Missing required MemberID (binding will fail if uuid.Nil is considered missing/invalid depending on binding, but uuid is struct so it is valid JSON. Wait, binding:required on UUID might fail if empty? Actually uuid.UUID is [16]byte, it's never empty. But `binding:"required"` for value types usually checks zero value. Let's assume sending empty json or invalid json tests ShouldBindJSON)
			},
			want: want{
				err:     true,
				code:    app.CodeBadRequest,
				Message: app.MessageBadRequest,
				data: wrapper.ResponseOption[member.GetMemberResponse]{
					HTTPStatus: http.StatusBadRequest,
					Code:       app.CodeBadRequest,
					Message:    app.MessageBadRequest,
				},
			},
		},
		{
			name: "fail, case GetMemberById error",
			prepare: func(m mockArgs, args args) {
				m.memberStorage.
					EXPECT().
					GetMemberById(args.ctx, args.req.MemberID).
					Return(access.Member{}, errors.New("storage error"))
			},
			args: args{
				req: member.GetMemberRequest{
					MemberID: memberID,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[member.GetMemberResponse]{
					HTTPStatus: http.StatusInternalServerError,
					Code:       app.CodeInternalError,
					Message:    app.MessageInternalError,
				},
			},
		},
		{
			name: "fail, case Decrypt error",
			prepare: func(m mockArgs, args args) {
				m.memberStorage.
					EXPECT().
					GetMemberById(args.ctx, args.req.MemberID).
					Return(access.Member{
						MemberID:       memberID.String(),
						Username:       "Test User",
						EncryptedEmail: "encrypted_email",
						HashedEmail:    "hashed_email",
						Status:         access.MemberStatusActive,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)

				m.aesgcm.
					EXPECT().
					Decrypt("encrypted_email").
					Return("", errors.New("decrypt error"))
			},
			args: args{
				req: member.GetMemberRequest{
					MemberID: memberID,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[member.GetMemberResponse]{
					HTTPStatus: http.StatusInternalServerError,
					Code:       app.CodeInternalError,
					Message:    app.MessageInternalError,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			// Arrange
			// For invalid body test, we handle it specially or just rely on struct zero value if binding catches it.
			// Ideally we test ShouldBindJSON failure by sending malformed JSON, but here we use struct.
			// Let's rely on `binding:"required"` logic.

			var payload bytes.Buffer
			if tt.name == "fail, case invalid request body" {
				payload.WriteString(`{}`) // Empty JSON
			} else {
				json.NewEncoder(&payload).Encode(tt.args.req)
			}

			req := httptest.NewRequest(http.MethodPost, "http://0.0.0.0/api/v1/platform/member/info", &payload)
			req.Header.Set("Content-Type", "application/json")

			ctx.Request = req

			m := mockArgs{
				memberStorage: access_mocks.NewMemberStorageMock(t),
				aesgcm:        aesgcm_mocks.NewAesgcmMock(t),
			}

			if tt.prepare != nil {
				tt.args.ctx = ctx.Request.Context()
				tt.prepare(m, tt.args)
			}

			h := member.NewHandler(member.HandlerConfig{
				MemberStorage: m.memberStorage,
				Aesgcm:        m.aesgcm,
				// Hash not used in GetInfo
				Hash: nil,
			})

			// Act
			h.GetInfo(ctx)

			// Assert
			var resp wrapper.ResponseOption[member.GetMemberResponse]
			json.NewDecoder(w.Body).Decode(&resp)

			if tt.want.err {
				r.NotEqual(http.StatusOK, w.Code)
				r.Equal(tt.want.code, resp.Code)
				r.Equal(tt.want.Message, resp.Message)
			} else {
				r.Equal(http.StatusOK, w.Code)
				r.Equal(tt.want.code, resp.Code)
				r.Equal(tt.want.Message, resp.Message)
				r.NotNil(resp.Data)
				r.Equal(memberID, resp.Data.MemberID)
			}
		})
	}
}
