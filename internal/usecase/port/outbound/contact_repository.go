package outbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type ContactRepository interface {
	SaveContact(ctx context.Context, contact domain.Contact) error
	GetContact(ctx context.Context, id string) (domain.Contact, error)
	GetContactView(ctx context.Context, id string) (domain.ContactView, error)
	ListContacts(ctx context.Context, filter domain.ContactFilter) ([]domain.ContactView, error)
	SaveOpportunityContact(ctx context.Context, link domain.OpportunityContact) error
	DeleteOpportunityContact(ctx context.Context, opportunityID, contactID string) error
	ListOpportunityContacts(ctx context.Context, opportunityID string) ([]domain.OpportunityContactView, error)
}
