package plugin

import (
	"testing"
)

func TestGRPCClient_App(t *testing.T) {
	client, _ := TestPluginGRPCConn(t, map[string]Plugin{
		"test": new(testInterfacePlugin),
	})
	defer client.Close()

	raw, err := client.Dispense("test")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	impl, ok := raw.(testInterface)
	if !ok {
		t.Fatalf("bad: %#v", raw)
	}

	result := impl.Double(21)
	if result != 42 {
		t.Fatalf("bad: %#v", result)
	}
}
