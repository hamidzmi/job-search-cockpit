package inbound

import (
	"context"

	"job-search-cockpit/internal/domain"
)

type ContactUseCase interface {
	CreateContact(ctx context.Context, in domain.ContactInput) (domain.ContactView, error)
	GetContact(ctx context.Context, id string) (domain.ContactView, error)
	ListContacts(ctx context.Context, filter domain.ContactFilter) ([]domain.ContactView, error)
	UpdateContact(ctx context.Context, id string, in domain.ContactInput) (domain.ContactView, error)
	LinkContactToOpportunity(ctx context.Context, in domain.OpportunityContactInput) (domain.OpportunityContactView, error)
	UnlinkContactFromOpportunity(ctx context.Context, opportunityID, contactID string) error
	ListOpportunityContacts(ctx context.Context, opportunityID string) ([]domain.OpportunityContactView, error)
}
