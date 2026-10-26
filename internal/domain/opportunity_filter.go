package domain

import "fmt"

type OpportunityFilter struct {
	Stage        Stage
	Outcome      Outcome
	Priority     Priority
	RemotePolicy RemotePolicy
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
	if f.Priority != "" && !f.Priority.IsValid() {
		return fmt.Errorf("%w: invalid priority filter %q", ErrValidation, f.Priority)
	}
	if f.RemotePolicy != "" && !f.RemotePolicy.IsValid() {
		return fmt.Errorf("%w: invalid remote policy filter %q", ErrValidation, f.RemotePolicy)
	}
	return nil
}
