package validate_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"awo.so/framework/def"
	"awo.so/framework/validate"
)

// ── test double ──────────────────────────────────────────────────────────────

type rec struct{ data map[string]any }

func newRec(kv ...any) *rec {
	r := &rec{data: make(map[string]any)}
	for i := 0; i+1 < len(kv); i += 2 {
		r.data[kv[i].(string)] = kv[i+1]
	}
	return r
}

func (r *rec) Get(f string) any    { return r.data[f] }
func (r *rec) Set(f string, v any) { r.data[f] = v }
func (r *rec) ID() uuid.UUID       { return uuid.Nil }
func (r *rec) TenantID() uuid.UUID { return uuid.Nil }
func (r *rec) EntityName() string  { return "test" }

var _ def.MutableRecord = (*rec)(nil)

// ── suite ────────────────────────────────────────────────────────────────────

type ValidateSuite struct{ suite.Suite }

func TestValidateSuite(t *testing.T) { suite.Run(t, new(ValidateSuite)) }

// ── built-in validators ───────────────────────────────────────────────────────

func (s *ValidateSuite) TestEmail_Valid() {
	s.Nil(validate.Email()("user@example.com", nil))
}

func (s *ValidateSuite) TestEmail_Invalid() {
	s.NotNil(validate.Email()("bad-email", nil))
}

func (s *ValidateSuite) TestEmail_Empty_Passes() {
	s.Nil(validate.Email()("", nil), "empty value must skip validation")
}

func (s *ValidateSuite) TestPhone_Valid() {
	s.Nil(validate.Phone()("+254700000000", nil))
}

func (s *ValidateSuite) TestPhone_MissingCountryCode() {
	s.NotNil(validate.Phone()("0700000000", nil))
}

func (s *ValidateSuite) TestURL_Valid() {
	s.Nil(validate.URL()("https://example.com", nil))
}

func (s *ValidateSuite) TestURL_Invalid() {
	s.NotNil(validate.URL()("not-a-url", nil))
}

func (s *ValidateSuite) TestURL_HTTP_Valid() {
	s.Nil(validate.URL()("http://example.com/path", nil))
}

func (s *ValidateSuite) TestKRAPin_Valid() {
	s.Nil(validate.KRAPin()("A000000000X", nil))
}

func (s *ValidateSuite) TestKRAPin_Invalid() {
	s.NotNil(validate.KRAPin()("123456789X", nil))
}

func (s *ValidateSuite) TestNHIF_Valid() {
	s.Nil(validate.NHIF()("123456", nil))
}

func (s *ValidateSuite) TestNHIF_TooShort() {
	s.NotNil(validate.NHIF()("123", nil))
}

func (s *ValidateSuite) TestMinLen_Sufficient() {
	s.Nil(validate.MinLen(5)("hello", nil))
}

func (s *ValidateSuite) TestMinLen_TooShort() {
	s.NotNil(validate.MinLen(5)("hi", nil))
}

func (s *ValidateSuite) TestRegex_Match() {
	s.Nil(validate.Regex(`^\d{4}$`)("1234", nil))
}

func (s *ValidateSuite) TestRegex_NoMatch() {
	s.NotNil(validate.Regex(`^\d{4}$`)("abc", nil))
}

// ── Run pipeline ──────────────────────────────────────────────────────────────

func (s *ValidateSuite) TestRun_Required_Missing() {
	def := &def.EntityDefinition{
		Name:   "invoice",
		Fields: []*def.FieldDef{def.String("title").Required()},
	}
	errs := validate.Run(def, newRec())
	s.Require().NotNil(errs)
	s.Len(errs, 1)
	s.Equal("title", errs[0].Field)
}

func (s *ValidateSuite) TestRun_Required_Present() {
	def := &def.EntityDefinition{
		Name:   "invoice",
		Fields: []*def.FieldDef{def.String("title").Required()},
	}
	s.Nil(validate.Run(def, newRec("title", "INV-001")))
}

func (s *ValidateSuite) TestRun_MaxLength_Exceeded() {
	def := &def.EntityDefinition{
		Name:   "invoice",
		Fields: []*def.FieldDef{def.String("note").MaxLen(5)},
	}
	errs := validate.Run(def, newRec("note", "toolongvalue"))
	s.NotNil(errs)
	s.Equal("note", errs[0].Field)
}

func (s *ValidateSuite) TestRun_Select_ValidOption() {
	def := &def.EntityDefinition{
		Name:   "invoice",
		Fields: []*def.FieldDef{def.Enum("status").Values("draft", "approved")},
	}
	s.Nil(validate.Run(def, newRec("status", "draft")))
}

func (s *ValidateSuite) TestRun_Select_InvalidOption() {
	def := &def.EntityDefinition{
		Name:   "invoice",
		Fields: []*def.FieldDef{def.Enum("status").Values("draft", "approved")},
	}
	errs := validate.Run(def, newRec("status", "unknown"))
	s.Require().NotNil(errs)
	s.Equal("status", errs[0].Field)
}

func (s *ValidateSuite) TestRun_RequiredFails_SkipsCustomValidator() {
	called := false
	def := &def.EntityDefinition{
		Name: "t",
		Fields: []*def.FieldDef{
			def.String("email").Required().Validate(func(v any, _ def.Record) *def.FieldError {
				called = true
				return nil
			}),
		},
	}
	errs := validate.Run(def, newRec())
	s.Require().NotNil(errs)
	s.False(called, "custom validator must not run when required check fails")
}

func (s *ValidateSuite) TestRun_EntityValidator_Runs_After_Fields() {
	crossCalled := false
	def := &def.EntityDefinition{
		Name:   "leave",
		Fields: []*def.FieldDef{def.String("start"), def.String("end")},
		EntityValidators: []def.EntityValidator{
			func(r def.Record) []*def.FieldError {
				crossCalled = true
				start, _ := r.Get("start").(string)
				end, _ := r.Get("end").(string)
				if start != "" && end != "" && end < start {
					return []*def.FieldError{{Field: "end", Message: "must be after start"}}
				}
				return nil
			},
		},
	}
	errs := validate.Run(def, newRec("start", "2026-06-10", "end", "2026-06-01"))
	s.True(crossCalled)
	s.Require().NotNil(errs)
	s.Equal("end", errs[0].Field)
}

func (s *ValidateSuite) TestRun_EntityValidator_SkippedOnFieldErrors() {
	crossCalled := false
	def := &def.EntityDefinition{
		Name:   "leave",
		Fields: []*def.FieldDef{def.String("start").Required()},
		EntityValidators: []def.EntityValidator{
			func(_ def.Record) []*def.FieldError {
				crossCalled = true
				return nil
			},
		},
	}
	// start is missing → field error → entity validator must not run
	errs := validate.Run(def, newRec())
	s.NotNil(errs)
	s.False(crossCalled, "entity validator must be skipped when field errors exist")
}

func (s *ValidateSuite) TestRun_MultipleFields_AllErrorsCollected() {
	def := &def.EntityDefinition{
		Name: "t",
		Fields: []*def.FieldDef{
			def.String("a").Required(),
			def.String("b").Required(),
		},
	}
	errs := validate.Run(def, newRec())
	s.Require().NotNil(errs)
	s.Len(errs, 2, "all field errors must be collected before returning")
}
