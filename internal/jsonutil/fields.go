// Package jsonutil reads optional fields out of a decoded JSON object
// without failing the whole parse. yt-dlp's JSON output is loosely typed —
// fields are missing, null, or a number where a string is expected
// depending on the extractor — so every accessor returns a zero value
// instead of an error.
package jsonutil

import "encoding/json"

// Object is a JSON object whose values are still undecoded.
type Object = map[string]json.RawMessage

// Decode parses raw as a JSON object.
func Decode(raw []byte) (Object, error) {
	var obj Object
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// Bool returns the boolean at key, or false if it is absent or not a bool.
func Bool(m Object, key string) bool {
	raw, ok := m[key]
	if !ok {
		return false
	}
	var b bool
	if err := json.Unmarshal(raw, &b); err != nil {
		return false
	}
	return b
}

// String returns the string at key, or "" if it is absent or not a string.
func String(m Object, key string) string {
	raw, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return ""
	}
	return s
}

// Float returns the number at key, or 0 if it is absent or not a number.
func Float(m Object, key string) float64 {
	raw, ok := m[key]
	if !ok {
		return 0
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err != nil {
		return 0
	}
	return f
}

// Int returns the number at key truncated to an int.
func Int(m Object, key string) int {
	return int(Float(m, key))
}

// Int64 returns the number at key as an int64, accepting both integer and
// floating-point encodings (yt-dlp emits filesizes either way).
func Int64(m Object, key string) int64 {
	raw, ok := m[key]
	if !ok {
		return 0
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return int64(f)
	}
	return 0
}

// Objects returns the array of objects at key, or nil if it is absent or
// not an array of objects.
func Objects(m Object, key string) []Object {
	raw, ok := m[key]
	if !ok {
		return nil
	}
	var out []Object
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}
