package db

import "github.com/google/uuid"

func StringPtr(s string) *string {
	return &s
}

func UUIDPtr(u uuid.UUID) *uuid.UUID {
	return &u
}

// Helper function to parse string and return *uuid.UUID
func ParseUUIDPtr(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	u, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
