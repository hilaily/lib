package sshx

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	proxyHost    = ""
	proxyKey     = ""
	proxyKey2    = ""
	proxyKeyPass = ""

	dstHost = ""
	dstPass = ""
)

func init() {
	proxyHost = os.Getenv("JUMP_HOST")
	proxyKey = os.Getenv("JUMP_KEY")
	proxyKey2 = "~/.ssh/id_pfs2"
	proxyKeyPass = "123123"

	dstHost = os.Getenv("DEST_HOST")
	dstPass = os.Getenv("DEST_PASS")
}

func TestNewClient1(t *testing.T) {
	// direct
	c, err := New(dstHost, dstPass, "", WithPort(59224), WithKeyPass(""))
	assert.NoError(t, err)
	check(t, c)
}

func createFile(t *testing.T, filename string) {
	err := os.WriteFile(filename, []byte("test copy 1"), 0777)
	assert.NoError(t, err)
}

func removeFile(filename string) {
	os.RemoveAll(filename)
}

func TestRawRsync(t *testing.T) {
	f := "/tmp/test-rsync.txt"
	createFile(t, f)
	defer func() {
		removeFile(f)
	}()
	c, err := New(proxyHost, "", proxyKey)
	assert.NoError(t, err)

	err = c.RawRsync(f, "/tmp/")
	assert.NoError(t, err)

	c, err = New(dstHost, dstPass, "", WithJumpProxy(proxyHost, "", proxyKey))
	assert.NoError(t, err)
	err = c.RawRsync(f, "/tmp/")
	assert.NoError(t, err)
}

func TestRawSCP(t *testing.T) {
	c, err := New(proxyHost, "", proxyKey)
	assert.NoError(t, err)

	err = c.RawSCP("/tmp/flower", "/tmp/")
	assert.NoError(t, err)

	c, err = New(dstHost, dstPass, "", WithJumpProxy(proxyHost, "", proxyKey))
	assert.NoError(t, err)
	err = c.RawSCP("/tmp/flower", "/tmp/")
	assert.NoError(t, err)
}

func TestRawInteract(t *testing.T) {
	c, err := New(proxyHost, "", proxyKey)
	assert.NoError(t, err)

	err = c.RawInteract()
	assert.NoError(t, err)

	c, err = New(dstHost, dstPass, "", WithJumpProxy(proxyHost, "", proxyKey))
	assert.NoError(t, err)
	err = c.RawInteract()
	assert.NoError(t, err)
}

func TestInteract(t *testing.T) {
	c, err := New(proxyHost, "", proxyKey)
	assert.NoError(t, err)

	err = c.Interact()
	t.Logf("err: %v", err)
	assert.NoError(t, err)
}

func TestNewClient(t *testing.T) {
	// direct
	c, err := New(proxyHost, "", proxyKey)
	assert.NoError(t, err)
	check(t, c)

	// direct with pass
	c, err = New(proxyHost, "", proxyKey2, WithKeyPass(proxyKeyPass))
	assert.NoError(t, err)
	check(t, c)

	// jump
	c, err = New(dstHost, dstPass, "", WithJumpProxy(proxyHost, "", proxyKey))
	assert.NoError(t, err)
	check(t, c)

}

func check(t *testing.T, c *Client) {
	res, err := c.RunResult("ls")
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	t.Log(string(res))
}

func TestSCP(t *testing.T) {
	c, err := New(proxyHost, "", proxyKey)
	assert.NoError(t, err)
	err = os.WriteFile("/tmp/1.txt", []byte("test copy 1"), 0777)
	assert.NoError(t, err)
	err = c.Copy("/tmp/1.txt", "/root/", true)
	assert.NoError(t, err)
	err = c.Copy("/tmp/1.txt", "/root/", false)
	assert.NoError(t, err)
	os.RemoveAll("/tmp/1.txt")

	os.MkdirAll("/tmp/test_copy/copy2", 0777)
	os.WriteFile("/tmp/test_copy/1.txt", []byte("1.txt"), 0777)
	os.WriteFile("/tmp/test_copy/copy2/2.txt", []byte("2.txt"), 0777)
	err = c.Copy("/tmp/test_copy/", "/tmp/test_copy", true)
	assert.NoError(t, err)
	os.RemoveAll("/tmp/test_copy/")
}
