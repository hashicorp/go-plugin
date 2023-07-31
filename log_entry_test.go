// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugin

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseJson(t *testing.T) {
	keys := []string{"firstKey", "secondKey", "thirdKey"}

	input := []byte(
		fmt.Sprintf(
			`{"@level":"info","@message":"msg","@timestamp":"2023-07-28T17:50:47.333365+02:00","%s":"1","%s":"2","%s":"3"}`,
			keys[0],
			keys[1],
			keys[2],
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
				t.Fatalf("expected: %v\ngot: %v", keys[i], entry.KVPairs[i].Key)
			}
		}
	}

}

func TestGetOrderedKeys(t *testing.T) {
	hclogKeys := []string{"@level", "@message", "@timestamp"}
	customKeys := []string{"firstKey", "secondKey", "thirdKey"}

	input := []byte(
		fmt.Sprintf(
			`{"%s":"info","%s":"msg","%s":"2023-07-28T17:50:47.333365+02:00","%s":"1","%s":"2","%s":"3"}`,
			hclogKeys[0],
			hclogKeys[1],
			hclogKeys[2],
			customKeys[0],
			customKeys[1],
			customKeys[2],
		),
	)

	expectedKeys := append(hclogKeys, customKeys...)
	actualKeys := getOrderedKeys(input)

	if !reflect.DeepEqual(expectedKeys, actualKeys) {
		t.Fatalf("expected: %v\ngot: %v", expectedKeys, actualKeys)
	}
}
