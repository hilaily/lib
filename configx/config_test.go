package configx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	conf, err := NewFromFile("testdata/test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var first struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	if err := conf.Get("first", &first); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, first.Name, "ace")
	assert.Equal(t, first.Age, 18)
	t.Logf("first: %+v", first)

	var second []struct {
		Name string `yaml:"name"`
		Age  int    `yaml:"age"`
	}
	if err := conf.Sub("second").Unmarshal(&second); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, second[0].Name, "bob")
	assert.Equal(t, second[0].Age, 19)
	assert.Equal(t, second[1].Name, "cat")
	assert.Equal(t, second[1].Age, 20)
	t.Logf("second: %+v", second)
}
