package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
	"github.com/stretchr/testify/assert"
)

func TestSaltEventMarshalJSON(t *testing.T) {
	alterTime := time.Now().Truncate(time.Second) // Truncate fractional seconds
	data := custom.JSON{
		Data: map[string]any{
			"key": "value",
		},
	}

	event := SaltEvent{
		ID:        1,
		Tag:       "test_tag",
		Data:      data,
		AlterTime: &alterTime,
		MasterID:  "master_123",
	}

	expectedJSON := `{
		"id": 1,
		"tag": "test_tag",
		"data": {"key":"value"},
		"alter_time": "` + alterTime.Format(time.RFC3339) + `",
		"master_id": "master_123"
	}`

	output, err := json.Marshal(event)
	assert.NoError(t, err)

	var expected map[string]any
	var actual map[string]any

	err = json.Unmarshal([]byte(expectedJSON), &expected)
	assert.NoError(t, err)

	err = json.Unmarshal(output, &actual)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestSaltEventMarshalJSONEmptyData(t *testing.T) {
	event := SaltEvent{
		ID:       1,
		Tag:      "test_tag",
		Data:     custom.JSON{Data: map[string]any{}},
		MasterID: "master_123",
	}

	expectedJSON := `{
		"id": 1,
		"tag": "test_tag",
		"data": {},
		"alter_time": null,
		"master_id": "master_123"
	}`

	output, err := json.Marshal(event)
	assert.NoError(t, err)

	var expected map[string]any
	var actual map[string]any

	err = json.Unmarshal([]byte(expectedJSON), &expected)
	assert.NoError(t, err)

	err = json.Unmarshal(output, &actual)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestSaltEventTableName(t *testing.T) {
	event := SaltEvent{}
	assert.Equal(t, "salt_events", event.TableName())
}
