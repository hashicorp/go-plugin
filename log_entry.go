package plugin

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// logEntry is the JSON payload that gets sent to Stderr from the plugin to the host
type logEntry struct {
	Message   string        `json:"message"`
	Level     string        `json:"level"`
	Timestamp time.Time     `json:"timestamp"`
	KVPairs   []*logEntryKV `json:"kv_pairs"`
}

// logEntryKV is a key value pair within the Output payload
type logEntryKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// parseKVPairs transforms string inputs into []*logEntryKV
func parseKVPairs(kvs ...interface{}) ([]*logEntryKV, error) {
	var result []*logEntryKV
	if len(kvs)%2 != 0 {
		return nil, fmt.Errorf("kv slice needs to be even number, got %d", len(kvs))
	}
	for i := 0; i < len(kvs); i = i + 2 {
		var val string

		switch st := kvs[i+1].(type) {
		case string:
			val = st
		case int:
			val = strconv.FormatInt(int64(st), 10)
		case int64:
			val = strconv.FormatInt(int64(st), 10)
		case int32:
			val = strconv.FormatInt(int64(st), 10)
		case int16:
			val = strconv.FormatInt(int64(st), 10)
		case int8:
			val = strconv.FormatInt(int64(st), 10)
		case uint:
			val = strconv.FormatUint(uint64(st), 10)
		case uint64:
			val = strconv.FormatUint(uint64(st), 10)
		case uint32:
			val = strconv.FormatUint(uint64(st), 10)
		case uint16:
			val = strconv.FormatUint(uint64(st), 10)
		case uint8:
			val = strconv.FormatUint(uint64(st), 10)
		default:
			val = fmt.Sprintf("%v", st)
		}

		result = append(result, &logEntryKV{
			Key:   kvs[i].(string),
			Value: val,
		})
	}

	return result, nil
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

// Payload returns a payload given a message, level, and kv pairs/
// If unable to parse KVs, it sets and returns the error message
// as the payload.
func payload(message string, level string, kvs ...interface{}) string {
	pairs, err := parseKVPairs(kvs...)
	if err != nil {
		return fmt.Sprintf("Unable to parse kv: %s\n", err)
	}
	entry := &logEntry{
		Message:   message,
		Level:     level,
		Timestamp: time.Now(),
		KVPairs:   pairs,
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf("Unable to marshal output: %s\n", err)
	}

	// Appedn newline to payload
	return fmt.Sprintf("%s\n", payload)
}
