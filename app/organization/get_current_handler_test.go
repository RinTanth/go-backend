package organization_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RinTanth/go-backend/app/organization"
	"github.com/RinTanth/go-backend/app/organization/access"
	access_mocks "github.com/RinTanth/go-backend/app/organization/access/mocks"
	"github.com/RinTanth/go-common/app"
	"github.com/RinTanth/go-common/wrapper"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetCurrent(t *testing.T) {
	r := require.New(t)

	memberID := uuid.New()
	organizationID := uuid.New()
	now := time.Now()
	joinedAt := now.Add(-24 * time.Hour)

	type mockArgs struct {
		organizationStorage *access_mocks.OrganizationStorageMock
	}

	type args struct {
		ctx context.Context
		req organization.GetCurrentRequest
	}

	type want struct {
		err     bool
		code    app.Code
		Message app.Message
		data    wrapper.ResponseOption[organization.GetCurrentResponse]
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
					GetMemberOrganization(args.ctx, args.req.MemberID).
					Return(access.OrganizationMember{
						OrganizationID: organizationID.String(),
						MemberID:       memberID.String(),
						Role:           access.OrganizationMemberRoleAdmin,
						Status:         access.OrganizationMemberStatusActive,
						JoinedAt:       &joinedAt,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)

				m.organizationStorage.
					EXPECT().
					GetOrganizationByID(args.ctx, organizationID).
					Return(access.Organization{
						OrganizationID: organizationID.String(),
						Name:           "Test Org",
						Status:         access.OrganizationStatusActive,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)
			},
			args: args{
				req: organization.GetCurrentRequest{
					MemberID: memberID,
				},
			},
			want: want{
				err:     false,
				code:    app.CodeSuccess,
				Message: app.MessageSuccess,
				data: wrapper.ResponseOption[organization.GetCurrentResponse]{
					HTTPStatus: http.StatusOK,
					Code:       app.CodeSuccess,
					Message:    app.MessageSuccess,
					Data: &organization.GetCurrentResponse{
						OrganizationID: organizationID,
						Name:           "Test Org",
						Status:         access.OrganizationStatusActive,
						Role:           access.OrganizationMemberRoleAdmin,
						JoinedAt:       &joinedAt,
						CreatedAt:      now,
						UpdatedAt:      now,
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
				req: organization.GetCurrentRequest{}, // Empty MemberID
			},
			want: want{
				err:     true,
				code:    app.CodeBadRequest,
				Message: app.MessageBadRequest,
				data: wrapper.ResponseOption[organization.GetCurrentResponse]{
					HTTPStatus: http.StatusBadRequest,
					Code:       app.CodeBadRequest,
					Message:    app.MessageBadRequest,
				},
			},
		},
		{
			name: "fail, case GetMemberOrganization error",
			prepare: func(m mockArgs, args args) {
				m.organizationStorage.
					EXPECT().
					GetMemberOrganization(args.ctx, args.req.MemberID).
					Return(access.OrganizationMember{}, errors.New("storage error"))
			},
			args: args{
				req: organization.GetCurrentRequest{
					MemberID: memberID,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[organization.GetCurrentResponse]{
					HTTPStatus: http.StatusInternalServerError,
					Code:       app.CodeInternalError,
					Message:    app.MessageInternalError,
				},
			},
		},
		{
			name: "fail, case GetOrganizationByID error",
			prepare: func(m mockArgs, args args) {
				m.organizationStorage.
					EXPECT().
					GetMemberOrganization(args.ctx, args.req.MemberID).
					Return(access.OrganizationMember{
						OrganizationID: organizationID.String(),
						MemberID:       memberID.String(),
						Role:           access.OrganizationMemberRoleAdmin,
						Status:         access.OrganizationMemberStatusActive,
						JoinedAt:       &joinedAt,
						CreatedAt:      now,
						UpdatedAt:      now,
					}, nil)

				m.organizationStorage.
					EXPECT().
					GetOrganizationByID(args.ctx, organizationID).
					Return(access.Organization{}, errors.New("org storage error"))
			},
			args: args{
				req: organization.GetCurrentRequest{
					MemberID: memberID,
				},
			},
			want: want{
				err:     true,
				code:    app.CodeInternalError,
				Message: app.MessageInternalError,
				data: wrapper.ResponseOption[organization.GetCurrentResponse]{
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
			if tt.name == "fail, case invalid request body" {
				payload.WriteString(`{}`)
			} else {
				json.NewEncoder(&payload).Encode(tt.args.req)
			}

			req := httptest.NewRequest(http.MethodPost, "http://0.0.0.0/api/v1/platform/organization/current", &payload)
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
			h.GetCurrent(ctx)

			// Assert
			var resp wrapper.ResponseOption[organization.GetCurrentResponse]
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
				r.Equal(organizationID, resp.Data.OrganizationID)
			}
		})
	}
}
