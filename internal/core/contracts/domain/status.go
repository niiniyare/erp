package domain

import "fmt"

// ContractStatus represents contract lifecycle states.
type ContractStatus string

const (
	StatusDraft           ContractStatus = "DRAFT"
	StatusPendingApproval ContractStatus = "PENDING_APPROVAL"
	StatusActive          ContractStatus = "ACTIVE"
	StatusExpired         ContractStatus = "EXPIRED"
	StatusTerminated      ContractStatus = "TERMINATED"
)

var validContractTransitions = map[ContractStatus][]ContractStatus{
	StatusDraft:           {StatusPendingApproval},
	StatusPendingApproval: {StatusActive, StatusDraft},
	StatusActive:          {StatusExpired, StatusTerminated},
	StatusExpired:         {},
	StatusTerminated:      {},
}

func (s ContractStatus) Valid() bool {
	_, ok := validContractTransitions[s]
	return ok
}

func (s ContractStatus) CanTransitionTo(target ContractStatus) bool {
	allowed, ok := validContractTransitions[s]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == target {
			return true
		}
	}
	return false
}

func (s ContractStatus) String() string { return string(s) }

func ParseContractStatus(s string) (ContractStatus, error) {
	st := ContractStatus(s)
	if !st.Valid() {
		return "", fmt.Errorf("invalid contract status: %q", s)
	}
	return st, nil
}

// ContractType classifies the nature of a contract.
type ContractType string

const (
	TypeVendor   ContractType = "VENDOR"
	TypeCustomer ContractType = "CUSTOMER"
	TypeEmployee ContractType = "EMPLOYEE"
	TypeService  ContractType = "SERVICE"
	TypeLease    ContractType = "LEASE"
	TypeOther    ContractType = "OTHER"
)

func (t ContractType) Valid() bool {
	switch t {
	case TypeVendor, TypeCustomer, TypeEmployee, TypeService, TypeLease, TypeOther:
		return true
	}
	return false
}
