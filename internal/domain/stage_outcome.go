package domain

type Stage string

const (
	StageDiscovered   Stage = "Discovered"
	StageOutreach     Stage = "Outreach"
	StageApplied      Stage = "Applied"
	StageScreening    Stage = "Screening"
	StageInterviewing Stage = "Interviewing"
	StageOffer        Stage = "Offer"
)

func AllStages() []Stage {
	return []Stage{
		StageDiscovered,
		StageOutreach,
		StageApplied,
		StageScreening,
		StageInterviewing,
		StageOffer,
	}
}

func (s Stage) IsValid() bool {
	switch s {
	case StageDiscovered, StageOutreach, StageApplied, StageScreening, StageInterviewing, StageOffer:
		return true
	default:
		return false
	}
}

type Outcome string

const (
	OutcomeActive    Outcome = "Active"
	OutcomeOnHold    Outcome = "On Hold"
	OutcomeWon       Outcome = "Won"
	OutcomeDeclined  Outcome = "Declined"
	OutcomeRejected  Outcome = "Rejected"
	OutcomeWithdrawn Outcome = "Withdrawn"
	OutcomeGhosted   Outcome = "Ghosted"
)

func AllOutcomes() []Outcome {
	return []Outcome{
		OutcomeActive,
		OutcomeOnHold,
		OutcomeWon,
		OutcomeDeclined,
		OutcomeRejected,
		OutcomeWithdrawn,
		OutcomeGhosted,
	}
}

func (o Outcome) IsValid() bool {
	switch o {
	case OutcomeActive, OutcomeOnHold, OutcomeWon, OutcomeDeclined, OutcomeRejected, OutcomeWithdrawn, OutcomeGhosted:
		return true
	default:
		return false
	}
}
