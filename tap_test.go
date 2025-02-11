package tap

import (
	"testing"

	"github.com/tj/assert"
)

func TestTap(t *testing.T) {
	res := Run(func(t *TAP) {
		t.Ok(true, "foo")
		t.Subtest("first subtest",
			func(t *TAP) {
				t.Plan(3)
				t.Ok(true, "bar")
				t.Ok(true, "baz")
			})
		t.Subtest("second subtest",
			func(t *TAP) {
				t.Plan(3)
				t.Diag("Hello")
				t.Ok(true, "bar")
				t.Ok(true, "baz")
				t.Subtest("third subtest",
					func(t *TAP) {
						t.Plan(3)
						t.Diag("World")
						t.Ok(true, "bar")
						t.Ok(true, "baz")
					})
			})
	})
	assert.Equal(t, res, false)

}
