package testsuite

import (
	"fmt"
	"testing"
)

// TestCase provides test utilities
type TestCase struct {
	t      *testing.T
	name   string
	assertions int
	failures int
}

// NewTestCase creates a new test case
func NewTestCase(t *testing.T, name string) *TestCase {
	fmt.Printf("\n=== Running: %s ===\n", name)
	return &TestCase{
		t:    t,
		name: name,
	}
}

// AssertEqual asserts that two values are equal
func (tc *TestCase) AssertEqual(expected, actual interface{}, msg string) {
	tc.assertions++
	if expected != actual {
		tc.failures++
		tc.t.Errorf("%s: Expected %v, got %v", msg, expected, actual)
	} else {
		fmt.Printf("  ✓ %s\n", msg)
	}
}

// AssertNotNil asserts that value is not nil
func (tc *TestCase) AssertNotNil(value interface{}, msg string) {
	tc.assertions++
	if value == nil {
		tc.failures++
		tc.t.Errorf("%s: Expected non-nil value", msg)
	} else {
		fmt.Printf("  ✓ %s\n", msg)
	}
}

// AssertTrue asserts that condition is true
func (tc *TestCase) AssertTrue(condition bool, msg string) {
	tc.assertions++
	if !condition {
		tc.failures++
		tc.t.Errorf("%s: Expected true", msg)
	} else {
		fmt.Printf("  ✓ %s\n", msg)
	}
}

// AssertFalse asserts that condition is false
func (tc *TestCase) AssertFalse(condition bool, msg string) {
	tc.assertions++
	if condition {
		tc.failures++
		tc.t.Errorf("%s: Expected false", msg)
	} else {
		fmt.Printf("  ✓ %s\n", msg)
	}
}

// AssertGreater asserts that value is greater than threshold
func (tc *TestCase) AssertGreater(value, threshold float64, msg string) {
	tc.assertions++
	if value <= threshold {
		tc.failures++
		tc.t.Errorf("%s: Expected %v > %v", msg, value, threshold)
	} else {
		fmt.Printf("  ✓ %s\n", msg)
	}
}

// Summary prints test summary
func (tc *TestCase) Summary() {
	fmt.Printf("\n=== Test Summary: %s ===\n", tc.name)
	fmt.Printf("  Assertions: %d\n", tc.assertions)
	fmt.Printf("  Failures: %d\n", tc.failures)
	fmt.Printf("  Result: %s\n", func() string {
		if tc.failures == 0 {
			return "PASSED"
		}
		return "FAILED"
	}())
}
