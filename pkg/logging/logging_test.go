package logging

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestIntializeDefaultLogger(t *testing.T) {
	assert.NotNil(t, InitializeDefaultLogger())
	assert.Equal(t, logrus.InfoLevel, InitializeDefaultLogger().Level)
}
func TestIntializeDefaultLoggerDebugtrue(t *testing.T) {
	os.Setenv("HAWKV6_GENERIC_PROCESSOR_DEBUG", "true")
	assert.NotNil(t, InitializeDefaultLogger())
	assert.Equal(t, logrus.DebugLevel, InitializeDefaultLogger().Level)
	os.Unsetenv("HAWKV6_GENERIC_PROCESSOR_DEBUG")
}

func TestIntializeDefaultLoggerDebugTRUE(t *testing.T) {
	os.Setenv("HAWKV6_GENERIC_PROCESSOR_DEBUG", "TRUE")
	assert.NotNil(t, InitializeDefaultLogger())
	assert.Equal(t, logrus.DebugLevel, InitializeDefaultLogger().Level)
	os.Unsetenv("HAWKV6_GENERIC_PROCESSOR_DEBUG")
}

func TestIntializeDefaultLoggerDebugNOTHING(t *testing.T) {
	os.Setenv("HAWKV6_GENERIC_PROCESSOR_DEBUG", "")
	assert.NotNil(t, InitializeDefaultLogger())
	assert.Equal(t, logrus.InfoLevel, InitializeDefaultLogger().Level)
	os.Unsetenv("HAWKV6_GENERIC_PROCESSOR_DEBUG")
}
