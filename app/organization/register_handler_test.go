package organization_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RinTanth/go-backend/app/organization"
	"github.com/RinTanth/go-backend/app/organization/access"
	access_mocks "github.com/RinTanth/go-backend/app/organization/access/mocks"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/wrapper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegisterOrganization(t *testing.T) {
	r := require.New(t)

	memberID := uuid.New()

	type mockArgs struct {
		organizationStorage *access_mocks.OrganizationStorageMock
	}

	type args struct {
		ctx context.Context
		req organization.RegisterOrganizationRequest
	}

	type want struct {
		err     bool
		code    app.Code
		Message app.Message
		data    wrapper.ResponseOption[organization.RegisterOrganizationResponse]
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
				m.organizationStorage.
					EXPECT().
					CreateOrganization(args.ctx, mock.Anything).
					Return(access.Organization{
						OrganizationID: "00000000-0000-0000-0000-000000000001",
					}, nil).
					Maybe()

				m.organizationStorage.
					EXPECT().
					CreateOrganizationMember(args.ctx, mock.Anything).
					Return(nil).
					Maybe()
			},
			args: args{
				req: organization.RegisterOrganizationRequest{
					Email:    "test@example.com",
					Name:     "Test User",
					MemberID: memberID,
				},
			},
			want: want{
				err:     false,
				code:    app.CodeSuccess,
				Message: app.MessageSuccess,
			},
		},
		{
			name: "fail, case missing email",
			prepare: func(m mockArgs, args args) {
				// no calls
			},
			args: args{
				req: organization.RegisterOrganizationRequest{
					Name:     "Test User",
					MemberID: memberID,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeBadRequest,
				Message: app.MessageBadRequest,
				data: wrapper.ResponseOption[organization.RegisterOrganizationResponse]{
					HTTPStatus: http.StatusBadRequest,
					Code:       app.CodeBadRequest,
					Message:    app.MessageBadRequest,
				},
			},
		},
		{
			name: "fail, case CreateOrganization error",
			prepare: func(m mockArgs, args args) {
				m.organizationStorage.
					EXPECT().
					CreateOrganization(args.ctx, mock.Anything).
					Return(access.Organization{}, errors.New("database error"))
			},
			args: args{
				req: organization.RegisterOrganizationRequest{
					Email:    "test@example.com",
					Name:     "Test User",
					MemberID: memberID,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[organization.RegisterOrganizationResponse]{
					HTTPStatus: http.StatusInternalServerError,
					Code:       app.CodeInternalError,
					Message:    app.MessageInternalError,
				},
			},
		},
		{
			name: "fail, case CreateOrganizationMember error",
			prepare: func(m mockArgs, args args) {
				m.organizationStorage.
					EXPECT().
					CreateOrganization(args.ctx, mock.Anything).
					Return(access.Organization{
						OrganizationID: "00000000-0000-0000-0000-000000000001",
					}, nil)

				m.organizationStorage.
					EXPECT().
					CreateOrganizationMember(args.ctx, mock.Anything).
					Return(errors.New("database error"))
			},
			args: args{
				req: organization.RegisterOrganizationRequest{
					Email:    "test@example.com",
					Name:     "Test User",
					MemberID: memberID,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[organization.RegisterOrganizationResponse]{
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

			req := httptest.NewRequest(http.MethodPost, "http://0.0.0.0/api/v1/platform/organization/register", &payload)
			req.Header.Set("Content-Type", "application/json")

			ctx.Request = req

			m := mockArgs{
				organizationStorage: access_mocks.NewOrganizationStorageMock(t),
			}

			if tt.prepare != nil {
				tt.args.ctx = ctx.Request.Context()
				tt.prepare(m, tt.args)
			}

			h := organization.NewHandler(organization.HandlerConfig{
				OrganizationStorage: m.organizationStorage,
			})

			// Act
			h.RegisterOrganization(ctx)

			// Assert
			var resp wrapper.ResponseOption[organization.RegisterOrganizationResponse]
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
				r.NotEqual(uuid.Nil, resp.Data.OrganizationID)
				r.Equal(memberID, resp.Data.MemberID)
				r.Equal(access.OrganizationMemberRoleAdmin, resp.Data.Role)
			}
		})
	}
}
