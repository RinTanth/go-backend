package member_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RinTanth/go-backend/app/member"
	"github.com/RinTanth/go-backend/app/member/access"
	access_mocks "github.com/RinTanth/go-backend/app/member/access/mocks"
	aesgcm_mocks "github.com/RinTanth/go-common/aesgcm/mocks"
	"github.com/RinTanth/go-common/app"
	hash_mocks "github.com/RinTanth/go-common/hash/mocks"
	"github.com/RinTanth/go-common/wrapper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterMember(t *testing.T) {
	r := require.New(t)

	type mockArgs struct {
		memberStorage *access_mocks.MemberStorageMock
		hash          *hash_mocks.HashManagerMock
		aesgcm        *aesgcm_mocks.AesgcmMock
	}

	type args struct {
		ctx context.Context
		req member.RegisterMemberRequest
	}

	type want struct {
		err     bool
		code    app.Code
		Message app.Message
		data    wrapper.ResponseOption[member.RegisterMemberResponse]
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
				m.hash.
					EXPECT().
					HashSha256EncodePepper(args.req.Email).
					Return("hashed_email")

				m.aesgcm.
					EXPECT().
					GenerateIV().
					Return([]byte("iv"), nil)

				m.aesgcm.
					EXPECT().
					Encrypt(args.req.Email, []byte("iv")).
					Return("encrypted_email", nil)

				m.memberStorage.
					EXPECT().
					CreateMember(args.ctx, mock.MatchedBy(func(input access.Member) bool {
						return input.Username == args.req.Name &&
							input.HashedEmail == "hashed_email" &&
							input.EncryptedEmail == "encrypted_email" &&
							input.Status == access.MemberStatusActive &&
							input.MemberID == "" &&
							input.CreatedAt.IsZero() &&
							input.UpdatedAt.IsZero()
					})).
					Return(access.Member{
						MemberID: "00000000-0000-0000-0000-000000000001",
					}, nil)
			},
			args: args{
				req: member.RegisterMemberRequest{
					Email: "test@example.com",
					Name:  "Test User",
				},
			},
			want: want{
				err:     false,
				code:    app.CodeSuccess,
				Message: app.MessageSuccess,
				data: wrapper.ResponseOption[member.RegisterMemberResponse]{
					HTTPStatus: http.StatusOK,
					Code:       app.CodeSuccess,
					Message:    app.MessageSuccess,
				},
			},
		},
		{
			name: "fail, case missing email",
			prepare: func(m mockArgs, args args) {
				// no calls
			},
			args: args{
				req: member.RegisterMemberRequest{
					Name: "Test User",
				},
			},
			want: want{
				err:     true,
				code:    app.CodeBadRequest,
				Message: app.MessageBadRequest,
				data: wrapper.ResponseOption[member.RegisterMemberResponse]{
					HTTPStatus: http.StatusBadRequest,
					Code:       app.CodeBadRequest,
					Message:    app.MessageBadRequest,
				},
			},
		},
		{
			name: "fail, case GenerateIV error",
			prepare: func(m mockArgs, args args) {
				m.hash.
					EXPECT().
					HashSha256EncodePepper(args.req.Email).
					Return("hashed_email")

				m.aesgcm.
					EXPECT().
					GenerateIV().
					Return(nil, errors.New("aesgcm error"))
			},
			args: args{
				req: member.RegisterMemberRequest{
					Email: "test@example.com",
					Name:  "Test User",
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[member.RegisterMemberResponse]{
					HTTPStatus: http.StatusInternalServerError,
					Code:       app.CodeInternalError,
					Message:    app.MessageInternalError,
				},
			},
		},
		{
			name: "fail, case Encrypt error",
			prepare: func(m mockArgs, args args) {
				m.hash.
					EXPECT().
					HashSha256EncodePepper(args.req.Email).
					Return("hashed_email")

				m.aesgcm.
					EXPECT().
					GenerateIV().
					Return([]byte("iv"), nil)

				m.aesgcm.
					EXPECT().
					Encrypt(args.req.Email, []byte("iv")).
					Return("", errors.New("encrypt error"))
			},
			args: args{
				req: member.RegisterMemberRequest{
					Email: "test@example.com",
					Name:  "Test User",
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[member.RegisterMemberResponse]{
					HTTPStatus: http.StatusInternalServerError,
					Code:       app.CodeInternalError,
					Message:    app.MessageInternalError,
				},
			},
		},
		{
			name: "fail, case CreateMember error",
			prepare: func(m mockArgs, args args) {
				m.hash.
					EXPECT().
					HashSha256EncodePepper(args.req.Email).
					Return("hashed_email")

				m.aesgcm.
					EXPECT().
					GenerateIV().
					Return([]byte("iv"), nil)

				m.aesgcm.
					EXPECT().
					Encrypt(args.req.Email, []byte("iv")).
					Return("encrypted_email", nil)

				m.memberStorage.
					EXPECT().
					CreateMember(args.ctx, mock.Anything).
					Return(access.Member{}, errors.New("storage error"))
			},
			args: args{
				req: member.RegisterMemberRequest{
					Email: "test@example.com",
					Name:  "Test User",
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[member.RegisterMemberResponse]{
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
			var payload bytes.Buffer
			json.NewEncoder(&payload).Encode(tt.args.req)

			req := httptest.NewRequest(http.MethodPost, "http://0.0.0.0/api/v1/platform/member/register", &payload)
			req.Header.Set("Content-Type", "application/json")

			ctx.Request = req

			m := mockArgs{
				memberStorage: access_mocks.NewMemberStorageMock(t),
				hash:          hash_mocks.NewHashManagerMock(t),
				aesgcm:        aesgcm_mocks.NewAesgcmMock(t),
			}

			if tt.prepare != nil {
				tt.args.ctx = ctx.Request.Context()
				tt.prepare(m, tt.args)
			}

			h := member.NewHandler(member.HandlerConfig{
				MemberStorage: m.memberStorage,
				Hash:          m.hash,
				Aesgcm:        m.aesgcm,
			})

			// Act
			h.RegisterMember(ctx)

			// Assert
			var resp wrapper.ResponseOption[member.RegisterMemberResponse]
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
				r.NotEqual(uuid.Nil, resp.Data.MemberID)
			}
		})
	}
}
