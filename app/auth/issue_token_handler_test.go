package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/wrapper"

	"github.com/RinTanth/go-backend/app/auth"

	token_mocks "github.com/RinTanth/go-common/token/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestIssueToken(t *testing.T) {

	r := require.New(t)

	userId := uuid.New().String()

	type mockArgs struct {
		token *token_mocks.JWTSignerMock
	}

	type args struct {
		ctx context.Context
		req auth.IssueTokenRequest
	}

	type want struct {
		err     bool
		code    app.Code
		Message app.Message
		data    wrapper.ResponseOption[auth.IssueTokenResponse]
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
				m.token.
					EXPECT().
					SignES256(mock.AnythingOfType("token.Claims")).
					Return("valid.access.token", nil)
			},
			args: args{
				req: auth.IssueTokenRequest{
					UserId: userId,
				},
			},
			want: want{
				err:     false,
				code:    app.CodeSuccess,
				Message: app.MessageSuccess,
				data: wrapper.ResponseOption[auth.IssueTokenResponse]{
					HTTPStatus: http.StatusOK,
					Code:       app.CodeSuccess,
					Message:    app.MessageSuccess,
					Data: &auth.IssueTokenResponse{
						AccessToken: "valid.access.token",
					},
				},
			},
		},
		{
			name: "fail, case invalid request body - missing user id",
			prepare: func(m mockArgs, args args) {
				// no token call
			},
			args: args{
				req: auth.IssueTokenRequest{
					// missing UserId
				},
			},
			want: want{
				err:     true,
				code:    app.CodeBadRequest,
				Message: app.MessageBadRequest,
				data: wrapper.ResponseOption[auth.IssueTokenResponse]{
					HTTPStatus: http.StatusBadRequest,
					Code:       app.CodeBadRequest,
					Message:    app.MessageBadRequest,
				},
			},
		},
		{
			name: "fail, case token signing error",
			prepare: func(m mockArgs, args args) {
				m.token.
					EXPECT().
					SignES256(mock.AnythingOfType("token.Claims")).
					Return("", assert.AnError)
			},
			args: args{
				req: auth.IssueTokenRequest{
					UserId: userId,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[auth.IssueTokenResponse]{
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

			req := httptest.NewRequest(http.MethodPost, "http://0.0.0.0/api/v1/auth/issue-token", &payload)
			req.Header.Set("Content-Type", "application/json")

			ctx.Request = req

			m := mockArgs{
				token: token_mocks.NewJWTSignerMock(t),
			}

			if tt.prepare != nil {
				tt.args.ctx = ctx.Request.Context()
				tt.prepare(m, tt.args)
			}

			h := auth.NewHandler(auth.HandlerConfig{
				Token: m.token,
			})

			// Act
			h.IssueToken(ctx)

			// Assert
			var resp wrapper.ResponseOption[auth.IssueTokenResponse]
			json.NewDecoder(w.Body).Decode(&resp)

			if tt.want.err {
				r.NotEqual(http.StatusOK, w.Code)
				r.Equal(tt.want.code, resp.Code)
				r.Equal(tt.want.Message, resp.Message)
			} else {
				r.Equal(http.StatusOK, w.Code)
				r.Equal(tt.want.code, resp.Code)
				r.Equal(tt.want.Message, resp.Message)
				r.Equal(tt.want.data.Data, resp.Data)
			}
		})
	}
}
