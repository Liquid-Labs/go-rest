package rest

import (
	"testing"

  "github.com/stretchr/testify/assert"
)

func TestMustGetEnv(t *testing.T) {
  assert.Panics(t, func() { MustGetenv(`NON_EXISTENT_ENV_VAR`) }, "MustGenEnv(`NON_EXISTENT_ENV_VAR`) failed to panic.")
   // TODO: justify assumption of 'PWD' or change (discover?)
  assert.NotPanics(t, func() { MustGetenv(`PWD`) }, "MustGenEnv(`PWD`) paniced.")
}

func TestGetEnvPurpose(t *testing.T) {
  assert.Equal(t, "test", GetEnvPurpose())
}
