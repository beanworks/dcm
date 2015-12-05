package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMapVal(t *testing.T) {
	fixture := map[interface{}]interface{}{
		"foo": "bar",
		"aaa": map[interface{}]interface{}{
			"bbb": "ccc",
		},
	}

	assert.NotPanics(t, func() { getMapVal(fixture, "foo") })
	assert.NotPanics(t, func() { getMapVal(fixture, "aaa", "bbb") })
	assert.Panics(t, func() { getMapVal(fixture, "invalid", "keys") })
	assert.Panics(t, func() { getMapVal(fixture, "foo", "bar", "baz", "qux") })

	assert.Equal(t, fixture, getMapVal(fixture))
	assert.Equal(t, "bar", getMapVal(fixture, "foo"))
	assert.Equal(t, "ccc", getMapVal(fixture, "aaa", "bbb"))
}
