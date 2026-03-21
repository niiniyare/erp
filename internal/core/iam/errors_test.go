package iam

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// ErrorsSuite — pure unit tests, no database required.
// ---------------------------------------------------------------------------

type ErrorsSuite struct{ suite.Suite }

func TestErrorsSuite(t *testing.T) { suite.Run(t, new(ErrorsSuite)) }

func (s *ErrorsSuite) TestErrorString_Format() {
	e := &Error{
		Code:       "AUTHZ_TEST",
		Message:    "test message",
		HTTPStatus: 418,
	}
	s.Equal("[authz] AUTHZ_TEST: test message", e.Error())
}

func (s *ErrorsSuite) TestSentinelErrors_Codes() {
	cases := []struct {
		err  *Error
		code string
	}{
		{ErrForbidden, "AUTHZ_FORBIDDEN"},
		{ErrUnauthorized, "AUTHZ_UNAUTHORIZED"},
		{ErrInvalidRequest, "AUTHZ_INVALID"},
		{ErrPolicyConflict, "AUTHZ_DUPLICATE"},
	}
	for _, tc := range cases {
		s.Run(tc.code, func() {
			s.Equal(tc.code, tc.err.Code)
		})
	}
}

func (s *ErrorsSuite) TestSentinelErrors_HTTPStatus() {
	s.Equal(403, ErrForbidden.HTTPStatus)
	s.Equal(401, ErrUnauthorized.HTTPStatus)
	s.Equal(400, ErrInvalidRequest.HTTPStatus)
	s.Equal(409, ErrPolicyConflict.HTTPStatus)
}

func (s *ErrorsSuite) TestSentinelErrors_NotNil() {
	s.NotNil(ErrForbidden)
	s.NotNil(ErrUnauthorized)
	s.NotNil(ErrInvalidRequest)
	s.NotNil(ErrPolicyConflict)
}

func (s *ErrorsSuite) TestSentinelErrors_ImplementsError() {
	// Verifies each sentinel satisfies the error interface.
	var _ error = ErrForbidden
	var _ error = ErrUnauthorized
	var _ error = ErrInvalidRequest
	var _ error = ErrPolicyConflict
}

func (s *ErrorsSuite) TestSentinelErrors_ErrorStringContainsMessage() {
	s.Contains(ErrForbidden.Error(), "access denied")
	s.Contains(ErrUnauthorized.Error(), "authentication required")
	s.Contains(ErrInvalidRequest.Error(), "required")
	s.Contains(ErrPolicyConflict.Error(), "already exists")
}
