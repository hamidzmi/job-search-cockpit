package domain

import (
	"fmt"
	"strings"
	"time"
)

type Company struct {
	ID           string
	Name         string
	Domain       string
	Headquarters string
	RemotePolicy string
	SizeStage    string
	Notes        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CompanyInput struct {
	Name         string
	Domain       string
	Headquarters string
	RemotePolicy string
	SizeStage    string
	Notes        string
}

func NewCompany(id string, in CompanyInput, now time.Time) (Company, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return Company{}, fmt.Errorf("%w: company name is required", ErrValidation)
	}
	if strings.TrimSpace(id) == "" {
		return Company{}, fmt.Errorf("%w: company id is required", ErrValidation)
	}
	return Company{
		ID:           strings.TrimSpace(id),
		Name:         name,
		Domain:       strings.TrimSpace(in.Domain),
		Headquarters: strings.TrimSpace(in.Headquarters),
		RemotePolicy: strings.TrimSpace(in.RemotePolicy),
		SizeStage:    strings.TrimSpace(in.SizeStage),
		Notes:        strings.TrimSpace(in.Notes),
		CreatedAt:    now.UTC(),
		UpdatedAt:    now.UTC(),
	}, nil
}

func (c *Company) Update(in CompanyInput, now time.Time) error {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return fmt.Errorf("%w: company name is required", ErrValidation)
	}
	c.Name = name
	c.Domain = strings.TrimSpace(in.Domain)
	c.Headquarters = strings.TrimSpace(in.Headquarters)
	c.RemotePolicy = strings.TrimSpace(in.RemotePolicy)
	c.SizeStage = strings.TrimSpace(in.SizeStage)
	c.Notes = strings.TrimSpace(in.Notes)
	c.UpdatedAt = now.UTC()
	return nil
}
