package dbtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSON map[string]any

func NewJSON(value map[string]any) JSON {
	if value == nil {
		return JSON{}
	}
	return JSON(value)
}

func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	data, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(data), nil
}

func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = JSON{}
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type JSON", value)
	}

	if len(data) == 0 {
		*j = JSON{}
		return nil
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	*j = JSON(result)
	return nil
}

func (j JSON) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(map[string]any(j))
}

func (j *JSON) UnmarshalJSON(data []byte) error {
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	*j = JSON(result)
	return nil
}
