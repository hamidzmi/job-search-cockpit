package domain

type Priority string

const (
	PriorityHigh   Priority = "High"
	PriorityMedium Priority = "Medium"
	PriorityLow    Priority = "Low"
)

func AllPriorities() []Priority {
	return []Priority{
		PriorityHigh,
		PriorityMedium,
		PriorityLow,
	}
}

func (p Priority) IsValid() bool {
	switch p {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return true
	default:
		return false
	}
}

type RemotePolicy string

const (
	RemotePolicyRemote RemotePolicy = "Remote"
	RemotePolicyHybrid RemotePolicy = "Hybrid"
	RemotePolicyOnsite RemotePolicy = "On-site"
)

func AllRemotePolicies() []RemotePolicy {
	return []RemotePolicy{
		RemotePolicyRemote,
		RemotePolicyHybrid,
		RemotePolicyOnsite,
	}
}

func (r RemotePolicy) IsValid() bool {
	switch r {
	case RemotePolicyRemote, RemotePolicyHybrid, RemotePolicyOnsite:
		return true
	default:
		return false
	}
}
