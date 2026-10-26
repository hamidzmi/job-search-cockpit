package domain

import "fmt"

type OpportunityFilter struct {
	Stage        Stage
	Outcome      Outcome
	Priority     string
	RemotePolicy string
	CompanyID    string
	Query        string
}

func (f OpportunityFilter) Validate() error {
	if f.Stage != "" && !f.Stage.IsValid() {
		return fmt.Errorf("%w: invalid stage filter %q", ErrValidation, f.Stage)
	}
	if f.Outcome != "" && !f.Outcome.IsValid() {
		return fmt.Errorf("%w: invalid outcome filter %q", ErrValidation, f.Outcome)
	}
	return nil
}
