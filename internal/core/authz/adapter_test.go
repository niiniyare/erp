package authz

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// ---------------------------------------------------------------------------
// AdapterHelpersSuite — tests for unexported helper functions in adapter.go.
// No database required.
// ---------------------------------------------------------------------------

type AdapterHelpersSuite struct{ suite.Suite }

func TestAdapterHelpersSuite(t *testing.T) { suite.Run(t, new(AdapterHelpersSuite)) }

// ---- ruleToValues ----------------------------------------------------------

func (s *AdapterHelpersSuite) TestRuleToValues_FullRule() {
	v0, v1, v2, v3, v4, v5 := ruleToValues([]string{"a", "b", "c", "d", "e", "f"})
	s.Equal("a", v0)
	s.Equal("b", v1)
	s.Equal("c", v2)
	s.Equal("d", v3)
	s.Equal("e", v4)
	s.Equal("f", v5)
}

func (s *AdapterHelpersSuite) TestRuleToValues_ShortRule_PadsWithEmpty() {
	v0, v1, v2, v3, v4, v5 := ruleToValues([]string{"sub", "dom", "obj"})
	s.Equal("sub", v0)
	s.Equal("dom", v1)
	s.Equal("obj", v2)
	s.Equal("", v3)
	s.Equal("", v4)
	s.Equal("", v5)
}

func (s *AdapterHelpersSuite) TestRuleToValues_EmptyRule_AllEmpty() {
	v0, v1, v2, v3, v4, v5 := ruleToValues([]string{})
	s.Equal("", v0)
	s.Equal("", v1)
	s.Equal("", v2)
	s.Equal("", v3)
	s.Equal("", v4)
	s.Equal("", v5)
}

func (s *AdapterHelpersSuite) TestRuleToValues_SingleElement() {
	v0, v1, v2, v3, v4, v5 := ruleToValues([]string{"only"})
	s.Equal("only", v0)
	s.Equal("", v1)
	s.Equal("", v2)
	s.Equal("", v3)
	s.Equal("", v4)
	s.Equal("", v5)
}

func (s *AdapterHelpersSuite) TestRuleToValues_FiveElements_PolicyRule() {
	// p = sub, dom, obj, act, eft
	v0, v1, v2, v3, v4, v5 := ruleToValues([]string{
		"role:finance-manager", "dom-1", "invoice/*", "*", "allow",
	})
	s.Equal("role:finance-manager", v0)
	s.Equal("dom-1", v1)
	s.Equal("invoice/*", v2)
	s.Equal("*", v3)
	s.Equal("allow", v4)
	s.Equal("", v5)
}

func (s *AdapterHelpersSuite) TestRuleToValues_DoesNotMutateInput() {
	input := []string{"a", "b", "c"}
	original := append([]string{}, input...) // copy
	ruleToValues(input)
	s.Equal(original, input, "input slice must not be modified")
}

// ---- filterEmpty -----------------------------------------------------------

func (s *AdapterHelpersSuite) TestFilterEmpty_RemovesTrailingEmpty() {
	result := filterEmpty([]string{"a", "b", "", ""})
	s.Equal([]string{"a", "b"}, result)
}

func (s *AdapterHelpersSuite) TestFilterEmpty_NoTrailingEmpty_Unchanged() {
	result := filterEmpty([]string{"a", "b", "c"})
	s.Equal([]string{"a", "b", "c"}, result)
}

func (s *AdapterHelpersSuite) TestFilterEmpty_AllEmpty_ReturnsEmpty() {
	result := filterEmpty([]string{"", "", ""})
	s.Len(result, 0)
}

func (s *AdapterHelpersSuite) TestFilterEmpty_EmptySlice_ReturnsEmpty() {
	result := filterEmpty([]string{})
	s.Len(result, 0)
}

func (s *AdapterHelpersSuite) TestFilterEmpty_InternalEmpty_Preserved() {
	// Only trailing empty strings are removed — internal ones stay.
	result := filterEmpty([]string{"a", "", "b", ""})
	s.Equal([]string{"a", "", "b"}, result)
}

func (s *AdapterHelpersSuite) TestFilterEmpty_SingleNonEmpty() {
	result := filterEmpty([]string{"only"})
	s.Equal([]string{"only"}, result)
}

func (s *AdapterHelpersSuite) TestFilterEmpty_SingleEmpty() {
	result := filterEmpty([]string{""})
	s.Len(result, 0)
}

// ---- joinRule --------------------------------------------------------------

func (s *AdapterHelpersSuite) TestJoinRule_MultipleParts() {
	result := joinRule([]string{"role:cfo", "dom", "invoice/*", "*", "allow"})
	s.Equal("role:cfo, dom, invoice/*, *, allow", result)
}

func (s *AdapterHelpersSuite) TestJoinRule_SinglePart() {
	result := joinRule([]string{"only"})
	s.Equal("only", result)
}

func (s *AdapterHelpersSuite) TestJoinRule_TwoParts() {
	result := joinRule([]string{"a", "b"})
	s.Equal("a, b", result)
}

func (s *AdapterHelpersSuite) TestJoinRule_EmptySlice() {
	result := joinRule([]string{})
	s.Equal("", result)
}

func (s *AdapterHelpersSuite) TestJoinRule_Separator() {
	// Delimiter must be ", " (comma space), matching persist.LoadPolicyLine format.
	result := joinRule([]string{"x", "y"})
	s.Contains(result, ", ")
	s.NotContains(result, " , ")
}
