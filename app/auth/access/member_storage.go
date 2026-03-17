package access

import (
	"context"
	"fmt"
	"time"

	gcpfirestore "cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type MemberStorage interface {
	GetMemberByEmail(ctx context.Context, hashedEmail string) (Member, error)
}

type memberStorage struct {
	fs *gcpfirestore.Client
}

var _ MemberStorage = (*memberStorage)(nil)

const memberCollection = "member"

func NewMemberStorage(fs *gcpfirestore.Client) MemberStorage {
	return &memberStorage{
		fs: fs,
	}
}

func (s *memberStorage) GetMemberByEmail(ctx context.Context, hashedEmail string) (Member, error) {
	if len(hashedEmail) == 0 {
		return Member{}, fmt.Errorf("hashed email is required")
	}

	docs, err := s.fs.Collection(memberCollection).
		Where("hashed_email", "==", hashedEmail).
		Limit(1).
		Documents(ctx).
		GetAll()
	if err != nil {
		return Member{}, fmt.Errorf("failed to query member: %w", err)
	}

	if len(docs) == 0 {
		return Member{}, fmt.Errorf("member not found")
	}

	var member Member
	if err := docs[0].DataTo(&member); err != nil {
		return Member{}, fmt.Errorf("failed to parse member data: %w", err)
	}

	return member, nil
}

type MemberStatusType string

var (
	MemberStatusActive    MemberStatusType = "ACTIVE"
	MemberStatusSuspended MemberStatusType = "SUSPENDED"
)

type Member struct {
	MemberID       string           `firestore:"member_id" json:"memberId"`
	Username       string           `firestore:"username" json:"username"`
	EncryptedEmail string           `firestore:"encrypted_email" json:"encryptedEmail"`
	HashedEmail    string           `firestore:"hashed_email" json:"hashedEmail"`
	Status         MemberStatusType `firestore:"status" json:"status"`
	CreatedAt      time.Time        `firestore:"created_at" json:"createdAt"`
	UpdatedAt      time.Time        `firestore:"updated_at" json:"updatedAt"`
	DeletedAt      *time.Time       `firestore:"deleted_at,omitempty" json:"deletedAt,omitempty"`
}

func (m *Member) GetID() uuid.UUID {
	return uuid.MustParse(m.MemberID)
}
