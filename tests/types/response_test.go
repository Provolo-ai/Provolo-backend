package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"provolo-api/internal/types"
)

func TestNewSuccessResponse(t *testing.T) {
	// Test with string data
	response := types.NewSuccessResponse("Test Title", "Test Message", "test data")
	
	assert.Equal(t, "Test Title", response.Title)
	assert.Equal(t, "Test Message", response.Message)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, "test data", response.Data)
}

func TestNewSuccessResponse_WithMapData(t *testing.T) {
	// Test with map data
	data := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}
	
	response := types.NewSuccessResponse("Map Title", "Map Message", data)
	
	assert.Equal(t, "Map Title", response.Title)
	assert.Equal(t, "Map Message", response.Message)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, data, response.Data)
}

func TestNewSuccessResponse_WithStructData(t *testing.T) {
	// Test with struct data
	type TestStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	
	data := TestStruct{
		Name:  "test",
		Value: 42,
	}
	
	response := types.NewSuccessResponse("Struct Title", "Struct Message", data)
	
	assert.Equal(t, "Struct Title", response.Title)
	assert.Equal(t, "Struct Message", response.Message)
	assert.Equal(t, "success", response.Status)
	assert.Equal(t, data, response.Data)
}

func TestNewSuccessResponse_WithNilData(t *testing.T) {
	// Test with nil data
	response := types.NewSuccessResponse("Nil Title", "Nil Message", nil)
	
	assert.Equal(t, "Nil Title", response.Title)
	assert.Equal(t, "Nil Message", response.Message)
	assert.Equal(t, "success", response.Status)
	assert.Nil(t, response.Data)
}

func TestNewErrorResponse(t *testing.T) {
	// Test error response
	response := types.NewErrorResponse("Error Title", "Error Message")
	
	assert.Equal(t, "Error Title", response.Title)
	assert.Equal(t, "Error Message", response.Message)
	assert.Equal(t, "error", response.Status)
	assert.Nil(t, response.Data)
}

func TestNewErrorResponse_EmptyStrings(t *testing.T) {
	// Test error response with empty strings
	response := types.NewErrorResponse("", "")
	
	assert.Equal(t, "", response.Title)
	assert.Equal(t, "", response.Message)
	assert.Equal(t, "error", response.Status)
	assert.Nil(t, response.Data)
}

func TestAPIResponse_JSONSerialization(t *testing.T) {
	// Test that APIResponse can be serialized to JSON
	response := types.NewSuccessResponse("JSON Test", "JSON Message", map[string]string{"key": "value"})
	
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	
	// Verify JSON structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	assert.NoError(t, err)
	
	assert.Equal(t, "JSON Test", parsed["title"])
	assert.Equal(t, "JSON Message", parsed["message"])
	assert.Equal(t, "success", parsed["status"])
	assert.NotNil(t, parsed["data"])
}

func TestAPIResponse_JSONDeserialization(t *testing.T) {
	// Test that APIResponse can be deserialized from JSON
	jsonData := `{
		"title": "Deserialized Title",
		"message": "Deserialized Message",
		"status": "success",
		"data": {"key": "value"}
	}`
	
	var response types.APIResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "Deserialized Title", response.Title)
	assert.Equal(t, "Deserialized Message", response.Message)
	assert.Equal(t, "success", response.Status)
	assert.NotNil(t, response.Data)
}

func TestAPIResponse_FieldTypes(t *testing.T) {
	// Test that all fields have the correct types
	response := types.APIResponse{
		Title:   "Test",
		Message: "Test Message",
		Status:  "success",
		Data:    "test data",
	}
	
	// Test field types
	assert.IsType(t, "", response.Title)
	assert.IsType(t, "", response.Message)
	assert.IsType(t, "", response.Status)
	assert.IsType(t, "", response.Data)
}

func TestNewSuccessResponse_Consistency(t *testing.T) {
	// Test that multiple calls with same parameters return consistent results
	response1 := types.NewSuccessResponse("Title", "Message", "data")
	response2 := types.NewSuccessResponse("Title", "Message", "data")
	
	assert.Equal(t, response1.Title, response2.Title)
	assert.Equal(t, response1.Message, response2.Message)
	assert.Equal(t, response1.Status, response2.Status)
	assert.Equal(t, response1.Data, response2.Data)
}

func TestNewErrorResponse_Consistency(t *testing.T) {
	// Test that multiple calls with same parameters return consistent results
	response1 := types.NewErrorResponse("Title", "Message")
	response2 := types.NewErrorResponse("Title", "Message")
	
	assert.Equal(t, response1.Title, response2.Title)
	assert.Equal(t, response1.Message, response2.Message)
	assert.Equal(t, response1.Status, response2.Status)
	assert.Equal(t, response1.Data, response2.Data)
}

func TestAPIResponse_EmptyValues(t *testing.T) {
	// Test with empty values
	response := types.APIResponse{
		Title:   "",
		Message: "",
		Status:  "",
		Data:    nil,
	}
	
	assert.Equal(t, "", response.Title)
	assert.Equal(t, "", response.Message)
	assert.Equal(t, "", response.Status)
	assert.Nil(t, response.Data)
}

func TestAPIResponse_JSONTags(t *testing.T) {
	// Test that JSON tags are correctly set
	response := types.NewSuccessResponse("Test", "Test", "data")
	
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	
	// Check that JSON keys match the struct tags
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	assert.NoError(t, err)
	
	// Verify all expected keys are present
	assert.Contains(t, parsed, "title")
	assert.Contains(t, parsed, "message")
	assert.Contains(t, parsed, "status")
	assert.Contains(t, parsed, "data")
}
