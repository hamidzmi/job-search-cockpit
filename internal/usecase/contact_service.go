package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"job-search-cockpit/internal/domain"
	"job-search-cockpit/internal/usecase/port/outbound"
)

type ContactService struct {
	tx        outbound.TxManager
	companies outbound.CompanyRepository
	opps      outbound.OpportunityRepository
	contacts  outbound.ContactRepository
	clock     outbound.Clock
}

func NewContactService(
	tx outbound.TxManager,
	companies outbound.CompanyRepository,
	opps outbound.OpportunityRepository,
	contacts outbound.ContactRepository,
	clock outbound.Clock,
) *ContactService {
	return &ContactService{
		tx:        tx,
		companies: companies,
		opps:      opps,
		contacts:  contacts,
		clock:     clock,
	}
}

func (s *ContactService) CreateContact(ctx context.Context, in domain.ContactInput) (domain.ContactView, error) {
	var created domain.ContactView
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		if companyID := strings.TrimSpace(in.CompanyID); companyID != "" {
			if _, err := s.companies.GetCompany(txCtx, companyID); err != nil {
				return err
			}
		}
		contact, err := domain.NewContact(uuid.NewString(), in, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.contacts.SaveContact(txCtx, contact); err != nil {
			return err
		}
		view, err := s.contacts.GetContactView(txCtx, contact.ID)
		if err != nil {
			return err
		}
		created = view
		return nil
	})
	if err != nil {
		return domain.ContactView{}, err
	}
	return created, nil
}

func (s *ContactService) GetContact(ctx context.Context, id string) (domain.ContactView, error) {
	return s.contacts.GetContactView(ctx, id)
}

func (s *ContactService) ListContacts(ctx context.Context, filter domain.ContactFilter) ([]domain.ContactView, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	return s.contacts.ListContacts(ctx, filter)
}

func (s *ContactService) UpdateContact(ctx context.Context, id string, in domain.ContactInput) (domain.ContactView, error) {
	var updated domain.ContactView
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		contact, err := s.contacts.GetContact(txCtx, id)
		if err != nil {
			return err
		}
		if companyID := strings.TrimSpace(in.CompanyID); companyID != "" {
			if _, err := s.companies.GetCompany(txCtx, companyID); err != nil {
				return err
			}
		}
		if err := contact.Update(in, s.clock.Now()); err != nil {
			return err
		}
		if err := s.contacts.SaveContact(txCtx, contact); err != nil {
			return err
		}
		view, err := s.contacts.GetContactView(txCtx, contact.ID)
		if err != nil {
			return err
		}
		updated = view
		return nil
	})
	if err != nil {
		return domain.ContactView{}, err
	}
	return updated, nil
}

func (s *ContactService) LinkContactToOpportunity(ctx context.Context, in domain.OpportunityContactInput) (domain.OpportunityContactView, error) {
	var linked domain.OpportunityContactView
	err := s.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		link, err := domain.NewOpportunityContact(in, s.clock.Now())
		if err != nil {
			return err
		}
		if _, err := s.opps.GetOpportunity(txCtx, link.OpportunityID); err != nil {
			return err
		}
		if _, err := s.contacts.GetContact(txCtx, link.ContactID); err != nil {
			return err
		}
		if err := s.contacts.SaveOpportunityContact(txCtx, link); err != nil {
			return err
		}
		views, err := s.contacts.ListOpportunityContacts(txCtx, link.OpportunityID)
		if err != nil {
			return err
		}
		for _, v := range views {
			if v.Contact.ID == link.ContactID {
				linked = v
				break
			}
		}
		return nil
	})
	if err != nil {
		return domain.OpportunityContactView{}, err
	}
	return linked, nil
}

func (s *ContactService) UnlinkContactFromOpportunity(ctx context.Context, opportunityID, contactID string) error {
	return s.contacts.DeleteOpportunityContact(ctx, opportunityID, contactID)
}

func (s *ContactService) ListOpportunityContacts(ctx context.Context, opportunityID string) ([]domain.OpportunityContactView, error) {
	if _, err := s.opps.GetOpportunity(ctx, opportunityID); err != nil {
		return nil, err
	}
	return s.contacts.ListOpportunityContacts(ctx, opportunityID)
}
