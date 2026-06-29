package sqlbuilder_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"awo.so/framework/filter"
	"awo.so/framework/persistence/sqlbuilder"
)

type FilterSQLSuite struct{ suite.Suite }

func TestFilterSQLSuite(t *testing.T) { suite.Run(t, new(FilterSQLSuite)) }

func (s *FilterSQLSuite) TestNil_NoOutput() {
	clause, args, next := sqlbuilder.ToSQL(nil, 1)
	s.Empty(clause)
	s.Nil(args)
	s.Equal(1, next)
}

func (s *FilterSQLSuite) TestNone_NoOutput() {
	clause, _, next := sqlbuilder.ToSQL(filter.None(), 1)
	s.Empty(clause)
	s.Equal(1, next)
}

func (s *FilterSQLSuite) TestEq() {
	clause, args, next := sqlbuilder.ToSQL(filter.Eq("status", "active"), 1)
	s.Equal("status = $1", clause)
	s.Equal([]any{"active"}, args)
	s.Equal(2, next)
}

func (s *FilterSQLSuite) TestNeq() {
	clause, args, _ := sqlbuilder.ToSQL(filter.Neq("status", "deleted"), 1)
	s.Contains(clause, "!=")
	s.Len(args, 1)
}

func (s *FilterSQLSuite) TestGt() {
	clause, _, _ := sqlbuilder.ToSQL(filter.Gt("amount", 100), 1)
	s.Contains(clause, ">")
	s.NotContains(clause, ">=")
}

func (s *FilterSQLSuite) TestGte() {
	clause, _, _ := sqlbuilder.ToSQL(filter.Gte("amount", 100), 1)
	s.Contains(clause, ">=")
}

func (s *FilterSQLSuite) TestLt() {
	clause, _, _ := sqlbuilder.ToSQL(filter.Lt("amount", 100), 1)
	s.Contains(clause, "<")
	s.NotContains(clause, "<=")
}

func (s *FilterSQLSuite) TestLte() {
	clause, _, _ := sqlbuilder.ToSQL(filter.Lte("amount", 100), 1)
	s.Contains(clause, "<=")
}

func (s *FilterSQLSuite) TestIn_Values() {
	clause, args, next := sqlbuilder.ToSQL(filter.In("status", "a", "b"), 1)
	s.Contains(clause, "IN")
	s.Len(args, 2)
	s.Equal(3, next)
}

func (s *FilterSQLSuite) TestIn_Empty_ReturnsFALSE() {
	clause, _, _ := sqlbuilder.ToSQL(filter.In("id"), 1)
	s.Equal("FALSE", clause)
}

func (s *FilterSQLSuite) TestNotIn_Empty_ReturnsTRUE() {
	clause, _, _ := sqlbuilder.ToSQL(filter.NotIn("id"), 1)
	s.Equal("TRUE", clause)
}

func (s *FilterSQLSuite) TestIsNull_NoArgs() {
	clause, args, next := sqlbuilder.ToSQL(filter.IsNull("deleted_at"), 1)
	s.Equal("deleted_at IS NULL", clause)
	s.Empty(args)
	s.Equal(1, next, "IsNull must not advance placeholder index")
}

func (s *FilterSQLSuite) TestIsNotNull() {
	clause, args, _ := sqlbuilder.ToSQL(filter.IsNotNull("approved_at"), 1)
	s.Contains(clause, "IS NOT NULL")
	s.Empty(args)
}

func (s *FilterSQLSuite) TestContains_WrapsPercent() {
	clause, args, _ := sqlbuilder.ToSQL(filter.Contains("name", "alice"), 1)
	s.Contains(clause, "ILIKE")
	s.Equal("%alice%", args[0])
}

func (s *FilterSQLSuite) TestStartsWith() {
	clause, args, _ := sqlbuilder.ToSQL(filter.StartsWith("code", "INV"), 1)
	s.Contains(clause, "LIKE")
	s.Equal("INV%", args[0])
}

func (s *FilterSQLSuite) TestEndsWith() {
	clause, args, _ := sqlbuilder.ToSQL(filter.EndsWith("code", "001"), 1)
	s.Contains(clause, "LIKE")
	s.Equal("%001", args[0])
}

func (s *FilterSQLSuite) TestAnd_CombinesChildren() {
	f := filter.Eq("a", 1).And(filter.Eq("b", 2))
	clause, args, next := sqlbuilder.ToSQL(f, 1)
	s.Contains(clause, "AND")
	s.Len(args, 2)
	s.Equal(3, next)
}

func (s *FilterSQLSuite) TestOr_CombinesChildren() {
	f := filter.Eq("x", 1).Or(filter.Eq("y", 2))
	clause, args, _ := sqlbuilder.ToSQL(f, 1)
	s.Contains(clause, "OR")
	s.Len(args, 2)
}

func (s *FilterSQLSuite) TestNot_Wraps() {
	f := filter.Not(filter.Eq("active", false))
	clause, args, _ := sqlbuilder.ToSQL(f, 1)
	s.True(strings.HasPrefix(clause, "NOT ("), "got: %s", clause)
	s.Len(args, 1)
}

func (s *FilterSQLSuite) TestStartIdx_PlaceholderOffset() {
	clause, _, next := sqlbuilder.ToSQL(filter.Eq("x", 99), 5)
	s.Contains(clause, "$5")
	s.Equal(6, next)
}

func (s *FilterSQLSuite) TestAnd_ThreeChildren() {
	f := filter.And(filter.Eq("a", 1), filter.Eq("b", 2), filter.Eq("c", 3))
	clause, args, next := sqlbuilder.ToSQL(f, 1)
	s.Contains(clause, "AND")
	s.Len(args, 3)
	s.Equal(4, next)
}
