// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package plugin

import (
	"bytes"
	"encoding/json"
	"sort"
	"time"
)

// logEntry is the JSON payload that gets sent to Stderr from the plugin to the host
type logEntry struct {
	Message   string        `json:"@message"`
	Level     string        `json:"@level"`
	Timestamp time.Time     `json:"timestamp"`
	KVPairs   []*logEntryKV `json:"kv_pairs"`
}

// logEntryKV is a key value pair within the Output payload
type logEntryKV struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// jsonKey has the log key and its position in the raw log. It's used for ordering
type jsonKey struct {
	key      string
	position int
}

// flattenKVPairs is used to flatten KVPair slice into []interface{}
// for hclog consumption.
func flattenKVPairs(kvs []*logEntryKV) []interface{} {
	var result []interface{}
	for _, kv := range kvs {
		result = append(result, kv.Key)
		result = append(result, kv.Value)
	}

	return result
}

// parseJSON handles parsing JSON output
func parseJSON(input []byte) (*logEntry, error) {
	var raw map[string]interface{}
	entry := &logEntry{}

	err := json.Unmarshal(input, &raw)
	if err != nil {
		return nil, err
	}

	// Parse hclog-specific objects
	if v, ok := raw["@message"]; ok {
		entry.Message = v.(string)
		delete(raw, "@message")
	}

	if v, ok := raw["@level"]; ok {
		entry.Level = v.(string)
		delete(raw, "@level")
	}

	if v, ok := raw["@timestamp"]; ok {
		t, err := time.Parse("2006-01-02T15:04:05.000000Z07:00", v.(string))
		if err != nil {
			return nil, err
		}
		entry.Timestamp = t
		delete(raw, "@timestamp")
	}

	orderedKeys := getOrderedKeys(input, raw)

	// Parse dynamic KV args from the hclog payload in order
	for _, k := range orderedKeys {
		entry.KVPairs = append(entry.KVPairs, &logEntryKV{
			Key:   k.key,
			Value: raw[k.key],
		})
	}

	return entry, nil
}

// getOrderedKeys returns the log keys ordered according to their original order of appearance
func getOrderedKeys(input []byte, raw map[string]interface{}) []jsonKey {
	var orderedKeys []jsonKey

	for key := range raw {
		position := bytes.Index(input, []byte("\""+key+"\":"))
		orderedKeys = append(orderedKeys, jsonKey{
			key,
			position,
		})
	}

	sort.Slice(orderedKeys, func(i, j int) bool {
		return orderedKeys[i].position < orderedKeys[j].position
	})

	return orderedKeys
}
