package filter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	t.Parallel()

	filter := Contains("field", "value")
	assert.Equal(t, "(cont,field,value)", filter.Filter())
}

func TestEquals(t *testing.T) {
	t.Parallel()

	filter := Equals("field", "value")
	assert.Equal(t, "(eq,field,value)", filter.Filter())
}

func TestGreaterThan(t *testing.T) {
	t.Parallel()

	filter := GreaterThan("field", "value")
	assert.Equal(t, "(gt,field,value)", filter.Filter())
}

func TestGreaterThanOrEqual(t *testing.T) {
	t.Parallel()

	filter := GreaterThanOrEqual("field", "value")
	assert.Equal(t, "(gte,field,value)", filter.Filter())
}

func TestIn(t *testing.T) {
	t.Parallel()

	filter := In("field", "value1", "value2")
	assert.Equal(t, "(in,field,value1,value2)", filter.Filter())
}

func TestLessThan(t *testing.T) {
	t.Parallel()

	filter := LessThan("field", "value")
	assert.Equal(t, "(lt,field,value)", filter.Filter())
}

func TestLessThanOrEqual(t *testing.T) {
	t.Parallel()

	filter := LessThanOrEqual("field", "value")
	assert.Equal(t, "(lte,field,value)", filter.Filter())
}

func TestDoesNotContain(t *testing.T) {
	t.Parallel()

	filter := DoesNotContain("field", "value")
	assert.Equal(t, "(ncont,field,value)", filter.Filter())
}

func TestDoesNotEqual(t *testing.T) {
	t.Parallel()

	filter := DoesNotEqual("field", "value")
	assert.Equal(t, "(neq,field,value)", filter.Filter())
}

func TestNotIn(t *testing.T) {
	t.Parallel()

	filter := NotIn("field", "value1", "value2")
	assert.Equal(t, "(nin,field,value1,value2)", filter.Filter())
}

func TestAnd(t *testing.T) {
	t.Parallel()

	filter1 := Equals("field1", "value1")
	filter2 := Contains("field2", "value2")
	andFilter := And(filter1, filter2)
	assert.Equal(t, "(eq,field1,value1);(cont,field2,value2)", andFilter.Filter())
}

func TestBasicFilterWithSpecialCharacters(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		value    string
		expected string
	}{
		{
			name:     "comma",
			value:    "value,with,comma",
			expected: "(eq,field,'value,with,comma')",
		},
		{
			name:     "space",
			value:    "value with space",
			expected: "(eq,field,'value with space')",
		},
		{
			name:     "parentheses",
			value:    "value(with)parentheses",
			expected: "(eq,field,'value(with)parentheses')",
		},
		{
			name:     "mixed special characters",
			value:    "value, with spaces (and) parentheses",
			expected: "(eq,field,'value, with spaces (and) parentheses')",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filter := Equals("field", tc.value)
			assert.Equal(t, tc.expected, filter.Filter())
		})
	}
}

func TestAndFilterWithMultipleFilters(t *testing.T) {
	t.Parallel()

	filter1 := Equals("field1", "value with spaces")
	filter2 := Contains("field2", "value,with,commas")
	filter3 := NotIn("field3", "value(with)parentheses", "normal_value")
	andFilter := And(filter1, filter2, filter3)

	expected := "(eq,field1,'value with spaces');" +
		"(cont,field2,'value,with,commas');" +
		"(nin,field3,'value(with)parentheses',normal_value)"
	assert.Equal(t, expected, andFilter.Filter())
}
