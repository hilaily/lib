package sshx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyID(t *testing.T) {
	c, err := New(dstHost, dstPass, "", WithJumpProxy(proxyHost, "", proxyKey))
	assert.NoError(t, err)
	err = c.CopyID()
	assert.NoError(t, err)
}
