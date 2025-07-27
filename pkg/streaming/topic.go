package streaming

import (
	"encoding/json"
	"fmt"
)

type Topic struct {
	Name   string
	Schema map[string]interface{}
}

func (t *Topic) Validate(payload []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("invalid json payload: %w", err)
	}

	for key, value := range t.Schema {
		if _, ok := data[key]; !ok {
			return fmt.Errorf("missing required field: %s", key)
		}
		// check type
		switch value.(type) {
		case string:
			if _, ok := data[key].(string); !ok {
				return fmt.Errorf("invalid type for field %s: expected string", key)
			}
		case float64:
			if _, ok := data[key].(float64); !ok {
				return fmt.Errorf("invalid type for field %s: expected number", key)
			}
		case bool:
			if _, ok := data[key].(bool); !ok {
				return fmt.Errorf("invalid type for field %s: expected boolean", key)
			}
		}
	}

	return nil
}
