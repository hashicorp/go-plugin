// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugin

import (
	"fmt"
	"testing"
)

func TestParseJson(t *testing.T) {
	keys := []string{"firstKey", "secondKey", "thirdKey"}
	raw := map[string]interface{}{
		keys[0]: "thirdKey", // we use keys as values to test correct key matching
		keys[1]: "secondKey",
		keys[2]: "firstKey",
	}

	input := []byte(
		fmt.Sprintf(
			`{"@level":"info","@message":"msg","@timestamp":"2023-07-28T17:50:47.333365+02:00","%s":"%s","%s":"%s","%s":"%s"}`,
			keys[0], raw[keys[0]],
			keys[1], raw[keys[1]],
			keys[2], raw[keys[2]],
		),
	)

	// the behavior is non deterministic, that's why this test is repeated multiple times
	for i := 0; i < 100; i++ {
		entry, err := parseJSON(input)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		for i := 0; i < len(keys); i++ {
			if keys[i] != entry.KVPairs[i].Key {
				t.Fatalf("expected key: %v\ngot key: %v", keys[i], entry.KVPairs[i].Key)
			}
			if raw[keys[i]] != entry.KVPairs[i].Value {
				t.Fatalf("expected value: %v\ngot value: %v", keys[i], entry.KVPairs[i].Key)
			}
		}
	}

}
