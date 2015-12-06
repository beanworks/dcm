package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMapVal(t *testing.T) {
	fixture := yamlConfig{
		"foo": "bar",
		"aaa": yamlConfig{
			"bbb": "ccc",
		},
	}

	assert.Equal(t, fixture, getMapVal(fixture))
	assert.Equal(t, "bar", getMapVal(fixture, "foo"))
	assert.Equal(t, yamlConfig{"bbb": "ccc"}, getMapVal(fixture, "aaa"))
	assert.Equal(t, "ccc", getMapVal(fixture, "aaa", "bbb"))

	assert.Equal(t, nil, getMapVal(fixture, "foo", "bar"))
	assert.Equal(t, nil, getMapVal(fixture, "foo", "bar", "baz"))
	assert.Equal(t, nil, getMapVal(fixture, "foo", "bar", "baz", "qux"))
	assert.Equal(t, nil, getMapVal(fixture, "aaa", "bbb", "ccc"))
	assert.Equal(t, nil, getMapVal(fixture, "aaa", "bbb", "ccc", "ddd"))
	assert.Equal(t, nil, getMapVal(fixture, "aaa", "bbb", "ccc", "ddd", "eee"))
	assert.Equal(t, nil, getMapVal(fixture, "invalid", "key"))
}
