package access

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepoer interface {
	GetUserByEmail(ctx context.Context, hashedEmail string) (Member, error)
}

type postgresRepo struct {
	pg *pgxpool.Pool
}

var _ PostgresRepoer = (*postgresRepo)(nil)

func NewPostgresRepo(pg *pgxpool.Pool) PostgresRepoer {
	return &postgresRepo{pg: pg}
}

func (r *postgresRepo) GetUserByEmail(ctx context.Context, hashedEmail string) (Member, error) {
	if len(hashedEmail) == 0 {
		return Member{}, fmt.Errorf("hashed email is required")
	}

	query := `
		SELECT member_id, username, encrypted_email, hashed_email, status, created_at, updated_at, deleted_at
		FROM member
		WHERE hashed_email = $1 AND deleted_at IS NULL
		LIMIT 1
	`

	var member Member
	err := r.pg.QueryRow(ctx, query, hashedEmail).Scan(
		&member.MemberID,
		&member.Username,
		&member.EncryptedEmail,
		&member.HashedEmail,
		&member.Status,
		&member.CreatedAt,
		&member.UpdatedAt,
		&member.DeletedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Member{}, fmt.Errorf("member not found")
		}
		return Member{}, fmt.Errorf("failed to query member: %w", err)
	}

	return member, nil
}

type MemberStatusType string

var (
	MemberStatusActive    MemberStatusType = "ACTIVE"
	MemberStatusSuspended MemberStatusType = "SUSPENDED"
)

type Member struct {
	MemberID       string           `db:"member_id" json:"memberId"`
	Username       string           `db:"username" json:"username"`
	EncryptedEmail string           `db:"encrypted_email" json:"encryptedEmail"`
	HashedEmail    string           `db:"hashed_email" json:"hashedEmail"`
	Status         MemberStatusType `db:"status" json:"status"`
	CreatedAt      time.Time        `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time        `db:"updated_at" json:"updatedAt"`
	DeletedAt      *time.Time       `db:"deleted_at" json:"deletedAt,omitempty"`
}

func (m *Member) GetID() uuid.UUID {
	return uuid.MustParse(m.MemberID)
}

type OrganizationMemberRoleType string

const (
	OrganizationMemberRoleAdmin OrganizationMemberRoleType = "ADMIN"
	OrganizationMemberRoleUser  OrganizationMemberRoleType = "USER"
)

type OrganizationMemberStatusType string

const (
	OrganizationMemberStatusActive  OrganizationMemberStatusType = "ACTIVE"
	OrganizationMemberStatusInvited OrganizationMemberStatusType = "INVITED"
	OrganizationMemberStatusRemoved OrganizationMemberStatusType = "REMOVED"
)

type OrganizationMember struct {
	ID             string                       `db:"id" json:"id"`
	OrganizationID string                       `db:"organization_id" json:"organizationId"`
	MemberID       string                       `db:"member_id" json:"memberId"`
	Role           OrganizationMemberRoleType   `db:"role" json:"role"`
	Status         OrganizationMemberStatusType `db:"status" json:"status"`
	JoinedAt       *time.Time                   `db:"joined_at" json:"joinedAt,omitempty"`
	CreatedAt      time.Time                    `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time                    `db:"updated_at" json:"updatedAt"`
	DeletedAt      *time.Time                   `db:"deleted_at" json:"deletedAt,omitempty"`
}

func (o *OrganizationMember) GetOrganizationID() uuid.UUID {
	return uuid.MustParse(o.OrganizationID)
}

func (o *OrganizationMember) GetMemberID() uuid.UUID {
	return uuid.MustParse(o.MemberID)
}
