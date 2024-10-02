package model

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSaltCacheMarshalJSON(t *testing.T) {
	alterTime := time.Now().Truncate(time.Second) // Truncate fractional seconds
	data := custom.JSON{
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	id := uuid.New()

	cache := SaltCache{
		Bank:      "test_bank",
		PSQLKey:   "test_key",
		Data:      data,
		ID:        id,
		AlterTime: &alterTime,
	}

	expectedJSON := `{
		"bank": "test_bank",
		"psql_key": "test_key",
		"data": {
			"key": "value"
		},
		"id": "` + id.String() + `",
		"alter_time": "` + alterTime.Format(time.RFC3339) + `"
	}`

	output, err := json.Marshal(cache)
	assert.NoError(t, err)

	var expected map[string]interface{}
	var actual map[string]interface{}

	err = json.Unmarshal([]byte(expectedJSON), &expected)
	assert.NoError(t, err)

	err = json.Unmarshal(output, &actual)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestSaltCacheMarshalJSONEmptyData(t *testing.T) {
	id := uuid.New()

	cache := SaltCache{
		Bank:    "test_bank",
		PSQLKey: "test_key",
		Data:    custom.JSON{Data: map[string]interface{}{}},
		ID:      id,
	}

	expectedJSON := `{
		"bank": "test_bank",
		"psql_key": "test_key",
		"data": {},
		"id": "` + id.String() + `",
		"alter_time": null
	}`

	output, err := json.Marshal(cache)
	assert.NoError(t, err)

	var expected map[string]interface{}
	var actual map[string]interface{}

	err = json.Unmarshal([]byte(expectedJSON), &expected)
	assert.NoError(t, err)

	err = json.Unmarshal(output, &actual)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}

func TestSaltCacheTableName(t *testing.T) {
	cache := SaltCache{}
	assert.Equal(t, "salt_cache", cache.TableName())
}
