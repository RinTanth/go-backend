package access

import (
	"context"
	"fmt"
	"time"

	gcpfirestore "cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type OrganizationStorage interface {
	GetOrganizationByID(ctx context.Context, organizationID uuid.UUID) (Organization, error)
	GetMemberOrganization(ctx context.Context, memberID uuid.UUID) (OrganizationMember, error)
	CreateOrganization(ctx context.Context, org Organization) (Organization, error)
	CreateOrganizationMember(ctx context.Context, orgMember OrganizationMember) error
}

type organizationStorage struct {
	fs *gcpfirestore.Client
}

var _ OrganizationStorage = (*organizationStorage)(nil)

const (
	organizationCollection       = "organization"
	organizationMemberCollection = "organization_member"
)

func NewOrganizationStorage(fs *gcpfirestore.Client) OrganizationStorage {
	return &organizationStorage{
		fs: fs,
	}
}

func (s *organizationStorage) GetOrganizationByID(ctx context.Context, organizationID uuid.UUID) (Organization, error) {
	if organizationID == uuid.Nil {
		return Organization{}, fmt.Errorf("organization id is required")
	}

	doc, err := s.fs.Collection(organizationCollection).
		Doc(organizationID.String()).
		Get(ctx)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to query organization: %w", err)
	}

	if doc == nil {
		return Organization{}, fmt.Errorf("organization not found")
	}

	var org Organization
	if err := doc.DataTo(&org); err != nil {
		return Organization{}, fmt.Errorf("failed to parse organization data: %w", err)
	}

	return org, nil
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

func (s *organizationStorage) CreateOrganization(ctx context.Context, org Organization) (Organization, error) {
	if org.OrganizationID == "" {
		org.OrganizationID = uuid.New().String()
	}

	now := time.Now().Local()
	if org.CreatedAt.IsZero() {
		org.CreatedAt = now
	}
	org.UpdatedAt = now

	_, err := s.fs.Collection(organizationCollection).Doc(org.OrganizationID).Set(ctx, org)
	if err != nil {
		return Organization{}, fmt.Errorf("failed to create organization: %w", err)
	}
	return org, nil
}

func (s *organizationStorage) CreateOrganizationMember(ctx context.Context, orgMember OrganizationMember) error {
	if orgMember.ID == "" {
		orgMember.ID = uuid.NewSHA1(orgMember.GetOrganizationID(), []byte(orgMember.MemberID)).String()
	}

	now := time.Now()
	orgMember.CreatedAt = now
	orgMember.UpdatedAt = now
	_, err := s.fs.Collection(organizationMemberCollection).Doc(orgMember.ID).Set(ctx, orgMember)
	if err != nil {
		return fmt.Errorf("failed to create organization member: %w", err)
	}
	return nil
}

type OrganizationStatusType string

const (
	OrganizationStatusActive   OrganizationStatusType = "ACTIVE"
	OrganizationStatusInactive OrganizationStatusType = "INACTIVE"
)

type Organization struct {
	ID             int64                  `firestore:"id" json:"id"`
	OrganizationID string                 `firestore:"organization_id" json:"organizationId"`
	Name           string                 `firestore:"name" json:"name"`
	Status         OrganizationStatusType `firestore:"status" json:"status"`
	CreatedAt      time.Time              `firestore:"created_at" json:"createdAt"`
	UpdatedAt      time.Time              `firestore:"updated_at" json:"updatedAt"`
	DeletedAt      *time.Time             `firestore:"deleted_at,omitempty" json:"deletedAt,omitempty"`
}

func (o *Organization) GetOrganizationID() uuid.UUID {
	return uuid.MustParse(o.OrganizationID)
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
