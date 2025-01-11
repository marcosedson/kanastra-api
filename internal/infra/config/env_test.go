package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_ExistingValue(t *testing.T) {
	err := os.Setenv("EXISTING_ENV", "test-value")
	if err != nil {
		return
	}

	defer func() {
		err := os.Unsetenv("EXISTING_ENV")
		if err != nil {

		}
	}()

	value := GetEnv("EXISTING_ENV", "default-value")

	assert.Equal(t, "test-value", value)
}

func TestGetEnv_DefaultValue(t *testing.T) {
	err := os.Unsetenv("NON_EXISTING_ENV")
	if err != nil {
		return
	}

	value := GetEnv("NON_EXISTING_ENV", "default-value")

	assert.Equal(t, "default-value", value)
}
