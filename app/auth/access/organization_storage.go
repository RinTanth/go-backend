package access

import (
	"context"
	"fmt"
	"time"

	gcpfirestore "cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type OrganizationStorage interface {
	GetMemberOrganization(ctx context.Context, memberID uuid.UUID) (OrganizationMember, error)
}

type organizationStorage struct {
	fs *gcpfirestore.Client
}

var _ OrganizationStorage = (*organizationStorage)(nil)

const organizationMemberCollection = "organization_member"

func NewOrganizationStorage(fs *gcpfirestore.Client) OrganizationStorage {
	return &organizationStorage{
		fs: fs,
	}
}

func (s *organizationStorage) GetMemberOrganization(ctx context.Context, memberID uuid.UUID) (OrganizationMember, error) {
	if memberID == uuid.Nil {
		return OrganizationMember{}, fmt.Errorf("member id is required")
	}

	docs, err := s.fs.Collection(organizationMemberCollection).
		Where("member_id", "==", memberID.String()).
		Limit(1).
		Documents(ctx).
		GetAll()
	if err != nil {
		return OrganizationMember{}, fmt.Errorf("failed to query organization member: %w", err)
	}

	if len(docs) == 0 {
		return OrganizationMember{}, fmt.Errorf("organization member not found")
	}

	var orgMember OrganizationMember
	if err := docs[0].DataTo(&orgMember); err != nil {
		return OrganizationMember{}, fmt.Errorf("failed to parse organization member data: %w", err)
	}

	return orgMember, nil
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
	ID             string                       `firestore:"id" json:"id"`
	OrganizationID string                       `firestore:"organization_id" json:"organizationId"`
	MemberID       string                       `firestore:"member_id" json:"memberId"`
	Role           OrganizationMemberRoleType   `firestore:"role" json:"role"`
	Status         OrganizationMemberStatusType `firestore:"status" json:"status"`
	JoinedAt       *time.Time                   `firestore:"joined_at,omitempty" json:"joinedAt,omitempty"`
	CreatedAt      time.Time                    `firestore:"created_at" json:"createdAt"`
	UpdatedAt      time.Time                    `firestore:"updated_at" json:"updatedAt"`
	DeletedAt      *time.Time                   `firestore:"deleted_at,omitempty" json:"deletedAt,omitempty"`
}

func (o *OrganizationMember) GetOrganizationID() uuid.UUID {
	return uuid.MustParse(o.OrganizationID)
}
func (o *OrganizationMember) GetMemberID() uuid.UUID {
	return uuid.MustParse(o.MemberID)
}
