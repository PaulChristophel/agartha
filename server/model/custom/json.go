// Package custom provides a JSON type that wraps around any to handle
// JSON encoding/decoding and database storage seamlessly. It includes implementations
// of the json.Marshaler, json.Unmarshaler, sql.Scanner, and driver.Valuer interfaces,
// allowing it to be used directly with JSON operations and SQL databases.

package custom

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSON is a wrapper around any that provides methods to handle JSON encoding/decoding
// and database storage. This allows the storage of arbitrary JSON data in a structured form
// that is compatible with Go's encoding/json package and database/sql package.
type JSON struct {
	Data any
}

// MarshalJSON implements the json.Marshaler interface.
// This method converts the JSON struct to a JSON-encoded byte slice.
func (j JSON) MarshalJSON() ([]byte, error) {
	// Marshal the Data field to JSON
	return json.Marshal(&j.Data)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// This method decodes a JSON-encoded byte slice and stores the result in the JSON struct.
func (j *JSON) UnmarshalJSON(data []byte) error {
	// Unmarshal the data into the Data field
	return json.Unmarshal(data, &j.Data)
}

// Value implements the driver.Valuer interface for database serialization.
// This method converts the JSON struct to a JSON-encoded byte slice for database storage.
func (j *JSON) Value() (driver.Value, error) {
	// Marshal the Data field to JSON for database storage
	return json.Marshal(&j.Data)
}

// Scan implements the sql.Scanner interface for database deserialization.
// This method scans a database value and decodes it into the JSON struct.
func (j *JSON) Scan(value any) error {
	var data []byte

	// Convert the database value to a byte slice
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("type assertion to []byte or string failed")
	}

	// Try to unmarshal the data into a temporary any
	var temp any
	if err := json.Unmarshal(data, &temp); err == nil {
		j.Data = temp
		return nil
	}

	// If unmarshalling fails, store the raw data as a string
	j.Data = string(data)
	return nil
}

// Ensure JSON implements the necessary interfaces.
var _ json.Marshaler = (*JSON)(nil)
var _ json.Unmarshaler = (*JSON)(nil)
var _ sql.Scanner = (*JSON)(nil)
var _ driver.Valuer = (*JSON)(nil)
