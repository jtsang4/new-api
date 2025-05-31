package dto

import (
	"encoding/json"
	"testing"
)

func TestFlexibleTimestamp_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name:     "integer timestamp",
			input:    `1748682323`,
			expected: 1748682323,
			wantErr:  false,
		},
		{
			name:     "floating-point timestamp (SambaNova case)",
			input:    `1748682323.3797884`,
			expected: 1748682323, // truncated to integer part
			wantErr:  false,
		},
		{
			name:     "string integer",
			input:    `"1748682323"`,
			expected: 1748682323,
			wantErr:  false,
		},
		{
			name:     "string float",
			input:    `"1748682323.3797884"`,
			expected: 1748682323,
			wantErr:  false,
		},
		{
			name:     "zero value",
			input:    `0`,
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "negative value",
			input:    `-123.456`,
			expected: -123,
			wantErr:  false,
		},
		{
			name:    "invalid string",
			input:   `"invalid"`,
			wantErr: true,
		},
		{
			name:     "null value",
			input:    `null`,
			expected: 0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ft FlexibleTimestamp
			err := json.Unmarshal([]byte(tt.input), &ft)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if ft.Int64() != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, ft.Int64())
			}
		})
	}
}

func TestOpenAITextResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{
			name: "integer created field",
			input: `{
				"id": "test-id",
				"model": "test-model",
				"object": "chat.completion",
				"created": 1748682323,
				"choices": [],
				"usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
			}`,
			expected: 1748682323,
			wantErr:  false,
		},
		{
			name: "floating-point created field (SambaNova case)",
			input: `{
				"id": "test-id",
				"model": "test-model",
				"object": "chat.completion",
				"created": 1748682323.3797884,
				"choices": [],
				"usage": {"prompt_tokens": 0, "completion_tokens": 0, "total_tokens": 0}
			}`,
			expected: 1748682323,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response OpenAITextResponse
			err := json.Unmarshal([]byte(tt.input), &response)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			
			if response.Created.Int64() != tt.expected {
				t.Errorf("expected created=%d, got %d", tt.expected, response.Created.Int64())
			}
		})
	}
}

func TestFlexibleTimestamp_MarshalJSON(t *testing.T) {
	ft := FlexibleTimestamp(1748682323)
	
	data, err := json.Marshal(ft)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	
	expected := `1748682323`
	if string(data) != expected {
		t.Errorf("expected %s, got %s", expected, string(data))
	}
}

func TestSambaNovaCompatibility(t *testing.T) {
	// This test simulates the exact SambaNova API response that was causing the error
	sambaNovaResponse := `{
		"id": "chatcmpl-test",
		"object": "chat.completion",
		"created": 1748682323.3797884,
		"model": "Meta-Llama-3.1-8B-Instruct",
		"choices": [
			{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you today?"
				},
				"finish_reason": "stop"
			}
		],
		"usage": {
			"prompt_tokens": 10,
			"completion_tokens": 8,
			"total_tokens": 18
		}
	}`

	var response OpenAITextResponse
	err := json.Unmarshal([]byte(sambaNovaResponse), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal SambaNova response: %v", err)
	}

	// Verify that the floating-point timestamp was correctly truncated to integer
	expectedCreated := int64(1748682323)
	if response.Created.Int64() != expectedCreated {
		t.Errorf("Expected created=%d, got %d", expectedCreated, response.Created.Int64())
	}

	// Verify other fields are correctly parsed
	if response.Id != "chatcmpl-test" {
		t.Errorf("Expected id='chatcmpl-test', got '%s'", response.Id)
	}

	if response.Model != "Meta-Llama-3.1-8B-Instruct" {
		t.Errorf("Expected model='Meta-Llama-3.1-8B-Instruct', got '%s'", response.Model)
	}

	// Test that the response can be marshaled back to JSON
	marshaledData, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response back to JSON: %v", err)
	}

	// Verify that the created field is marshaled as an integer
	var parsedBack map[string]interface{}
	err = json.Unmarshal(marshaledData, &parsedBack)
	if err != nil {
		t.Fatalf("Failed to parse marshaled JSON: %v", err)
	}

	createdValue, ok := parsedBack["created"].(float64)
	if !ok {
		t.Fatalf("Created field is not a number in marshaled JSON")
	}

	if int64(createdValue) != expectedCreated {
		t.Errorf("Marshaled created field: expected %d, got %d", expectedCreated, int64(createdValue))
	}
}
