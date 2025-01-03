package vstack_api

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// CodeUnion is a custom type designed to handle the "code" field in API responses,
// which can be either an integer or a string. This flexibility is necessary because
// different API endpoints might return the "code" in varying formats.
type CodeUnion struct {
	AsInt    *int64  // AsInt holds the integer representation of the code if available.
	AsString *string // AsString holds the string representation of the code if available.
}

// UnmarshalJSON implements the json.Unmarshaler interface for CodeUnion.
// It attempts to unmarshal the JSON data into an integer first.
// If that fails, it then tries to unmarshal it into a string.
// If both attempts fail, it returns an error.
func (u *CodeUnion) UnmarshalJSON(data []byte) error {
	// Attempt to unmarshal data into an int64.
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		u.AsInt = &i
		return nil
	}

	// If unmarshaling into int64 fails, attempt to unmarshal into a string.
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		u.AsString = &s
		return nil
	}

	// If both unmarshaling attempts fail, return an error.
	return fmt.Errorf("cannot unmarshal CodeUnion: not int nor string")
}

// CodeAsInt returns the code as an int64.
// If the code was originally an integer, it returns its value.
// If the code was a string that can be parsed into an integer, it returns the parsed value.
// If parsing fails or the code was a non-integer string, it returns 0.
func (u *CodeUnion) CodeAsInt() int64 {
	if u.AsInt != nil {
		return *u.AsInt
	}
	if u.AsString != nil {
		if val, err := strconv.ParseInt(*u.AsString, 10, 64); err == nil {
			return val
		}
	}
	return 0
}

// CodeAsString returns the code as a string.
// If the code was originally a string, it returns its value.
// If the code was an integer, it formats and returns it as a string.
// If neither is available, it returns an empty string.
func (u *CodeUnion) CodeAsString() string {
	if u.AsString != nil {
		return *u.AsString
	}
	if u.AsInt != nil {
		return fmt.Sprintf("%d", *u.AsInt)
	}
	return ""
}
