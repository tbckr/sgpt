package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestAWSBedrockProvider_NonStreaming(t *testing.T) {
	tests := []struct {
		name     string
		response map[string]interface{}
		want     string
	}{
		{
			name: "root level content (actual AWS response)",
			response: map[string]interface{}{
				"id":   "msg_bdrk_016f9z2E2X5RKKnCBcHoQugo",
				"type": "message",
				"role": "assistant",
				"content": []map[string]string{
					{
						"type": "text",
						"text": "The mass of the Sun is approximately 1.989 × 10^30 kilograms",
					},
				},
				"stop_reason": "end_turn",
				"usage": map[string]int{
					"input_tokens":  10,
					"output_tokens": 142,
				},
			},
			want: "The mass of the Sun is approximately 1.989 × 10^30 kilograms",
		},
		{
			name: "message nested content (test format)",
			response: map[string]interface{}{
				"type": "message",
				"message": map[string]interface{}{
					"content": []map[string]string{
						{
							"type": "text",
							"text": "The mass of the Sun is approximately 1.989 × 10^30 kilograms",
						},
					},
					"role": "assistant",
				},
			},
			want: "The mass of the Sun is approximately 1.989 × 10^30 kilograms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal response to bytes
			responseBytes, err := json.Marshal(tt.response)
			require.NoError(t, err)

			// Create config with test settings
			config := viper.New()
			config.Set("model", "us.anthropic.claude-3-5-sonnet-20241022-v2:0")
			config.Set("stream", false)
			config.Set("provider", "bedrock")
			config.Set("maxtokens", 1000)
			config.Set("temperature", 0.7)

			// Create output buffer to capture response
			var output strings.Builder

			// Create provider with mock client
			provider := &AWSBedrockProvider{
				config:   config,
				out:     &output,
				client:  createTestClient(responseBytes),
				testMode: true,
			}

			// Test non-streaming completion
			result, err := provider.CreateCompletion(context.Background(), "test", []string{"mass of sun"}, "", nil)
			require.NoError(t, err)
			require.Equal(t, tt.want, result)
			require.Contains(t, output.String(), tt.want)
		})
	}
}

func TestAWSBedrockProvider_Streaming(t *testing.T) {
	// Create mock streaming chunks
	chunks := [][]byte{
		[]byte(`{"type":"message_start","delta":{"message":{"content":[{"type":"text","text":"The mass"}]},"role":"assistant"}}`),
		[]byte(`{"type":"content_block_delta","delta":{"type":"text","text":" of the Sun"}}`),
		[]byte(`{"type":"content_block_delta","delta":{"type":"text","text":" is approximately"}}`),
		[]byte(`{"type":"content_block_delta","delta":{"type":"text","text":" 1.989 × 10^30 kilograms%"}}`), // Include trailing %
		[]byte(`{"type":"content_block_delta","delta":{"type":"text","text":" 1.989 × 10^30 kilograms%"}}`), // Repeat text to test deduplication
	}

	// Create mock client with streaming chunks
	mockClient := &testModeClient{
		mockResponse: chunks[0], // Use first chunk as mock response
	}

	// Create a buffer to capture streaming output
	var output strings.Builder

	// Create config with test settings
	config := viper.New()
	config.Set("model", "us.anthropic.claude-3-5-sonnet-20241022-v2:0")
	config.Set("stream", true)
	config.Set("provider", "bedrock")
	config.Set("maxtokens", 1000)
	config.Set("temperature", 0.7)

	// Create provider with mock client
	provider := &AWSBedrockProvider{
		config: config,
		out:    &output,
		client: mockClient,
		testMode: true,
	}

	// Test streaming completion
	result, err := provider.CreateCompletion(context.Background(), "test", []string{"mass of sun"}, "", nil)
	require.NoError(t, err)

	// Verify streaming output
	streamOutput := output.String()
	fmt.Printf("Stream output: %q\n", streamOutput)

	// Verify no text repetition
	require.Equal(t, "The mass of the Sun is approximately 1.989 × 10^30 kilograms", result)
	require.Equal(t, "The mass of the Sun is approximately 1.989 × 10^30 kilograms", streamOutput)
	
	// Verify no trailing % character
	require.NotContains(t, streamOutput, "%")
}

// createTestClient creates a new test client with the given response
func createTestClient(response []byte) BedrockInvoker {
	return &testModeClient{
		mockResponse: response,
	}
}
