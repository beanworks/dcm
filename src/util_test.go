package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMapValue(t *testing.T) {
	fixture := map[interface{}]interface{}{
		"foo": "bar",
		"aaa": map[interface{}]interface{}{
			"bbb": "ccc",
		},
	}

	assert.NotPanics(t, func() { getMapValue(fixture, "foo") })
	assert.NotPanics(t, func() { getMapValue(fixture, "aaa", "bbb") })
	assert.Panics(t, func() { getMapValue(fixture, "invalid", "keys") })
	assert.Panics(t, func() { getMapValue(fixture, "foo", "bar", "baz", "qux") })

	assert.Equal(t, fixture, getMapValue(fixture))
	assert.Equal(t, "bar", getMapValue(fixture, "foo"))
	assert.Equal(t, "ccc", getMapValue(fixture, "aaa", "bbb"))
}
