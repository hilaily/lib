package proxy

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrapSOCKSProxy(t *testing.T) {
	client := http.DefaultClient
	client, err := WrapSOCKSProxy(client,"139.162.78.109:8080",true)
	assert.NoError(t,err)
	resp, err := client.Get("http://amazon.com/")
	assert.NoError(t,err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK,resp.StatusCode)
	t.Logf("%v\n",resp)
}
