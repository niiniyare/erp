package filter_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"awo.so/framework/filter"
)

type FilterSuite struct{ suite.Suite }

func TestFilterSuite(t *testing.T) { suite.Run(t, new(FilterSuite)) }

func (s *FilterSuite) TestEq_Kind() {
	f := filter.Eq("name", "alice")
	s.Equal(filter.KindEq, f.Kind)
	s.Equal("name", f.Field)
	s.Equal("alice", f.Value)
}

func (s *FilterSuite) TestNeq() {
	f := filter.Neq("status", "deleted")
	s.Equal(filter.KindNeq, f.Kind)
}

func (s *FilterSuite) TestGt_Gte_Lt_Lte() {
	s.Equal(filter.KindGt, filter.Gt("amount", 0).Kind)
	s.Equal(filter.KindGte, filter.Gte("amount", 0).Kind)
	s.Equal(filter.KindLt, filter.Lt("amount", 0).Kind)
	s.Equal(filter.KindLte, filter.Lte("amount", 0).Kind)
}

func (s *FilterSuite) TestIn_Values() {
	f := filter.In("status", "active", "pending")
	s.Equal(filter.KindIn, f.Kind)
	s.Len(f.Values, 2)
}

func (s *FilterSuite) TestNotIn() {
	f := filter.NotIn("id", "x", "y", "z")
	s.Equal(filter.KindNotIn, f.Kind)
	s.Len(f.Values, 3)
}

func (s *FilterSuite) TestContains() {
	f := filter.Contains("name", "ali")
	s.Equal(filter.KindContains, f.Kind)
	s.Equal("ali", f.Value)
}

func (s *FilterSuite) TestStartsWith() {
	f := filter.StartsWith("code", "INV")
	s.Equal(filter.KindStartsWith, f.Kind)
}

func (s *FilterSuite) TestEndsWith() {
	f := filter.EndsWith("code", "001")
	s.Equal(filter.KindEndsWith, f.Kind)
}

func (s *FilterSuite) TestIsNull() {
	f := filter.IsNull("deleted_at")
	s.Equal(filter.KindIsNull, f.Kind)
	s.Equal("deleted_at", f.Field)
}

func (s *FilterSuite) TestIsNotNull() {
	f := filter.IsNotNull("approved_at")
	s.Equal(filter.KindIsNotNull, f.Kind)
}

func (s *FilterSuite) TestNone() {
	f := filter.None()
	s.Equal(filter.KindNone, f.Kind)
}

func (s *FilterSuite) TestAnd_MethodChain() {
	f := filter.Eq("a", 1).And(filter.Eq("b", 2))
	s.Equal(filter.KindAnd, f.Kind)
	s.Len(f.Children, 2)
}

func (s *FilterSuite) TestOr_MethodChain() {
	f := filter.Eq("a", 1).Or(filter.Eq("b", 2))
	s.Equal(filter.KindOr, f.Kind)
	s.Len(f.Children, 2)
}

func (s *FilterSuite) TestNot() {
	f := filter.Not(filter.Eq("active", true))
	s.Equal(filter.KindNot, f.Kind)
	s.NotNil(f.Inner)
	s.Equal(filter.KindEq, f.Inner.Kind)
}

func (s *FilterSuite) TestAnd_FunctionForm() {
	f := filter.And(filter.Eq("x", 1), filter.Eq("y", 2), filter.Eq("z", 3))
	s.Equal(filter.KindAnd, f.Kind)
	s.Len(f.Children, 3)
}

func (s *FilterSuite) TestOr_FunctionForm() {
	f := filter.Or(filter.Eq("x", 1), filter.Eq("y", 2))
	s.Equal(filter.KindOr, f.Kind)
	s.Len(f.Children, 2)
}
