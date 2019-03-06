package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	main()
	assert.Nil(t, nil)
	assert.NotNil(t, "string")
}
