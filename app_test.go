package plugin

import (
	"testing"
)

func TestApp(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("app")})
	defer c.Kill()

	_, err := c.Client()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
