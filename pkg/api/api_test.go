// Copyright (c) 2023 Tim <tbckr>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// SPDX-License-Identifier: MIT

package api

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"github.com/tbckr/sgpt/v2/internal/testlib"
	"github.com/tbckr/sgpt/v2/pkg/chat"
)

func TestCreateClient(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test")

	var client *OpenAIClient
	var err error
	client, err = CreateClient(nil, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientMissingApiKey(t *testing.T) {
	prev, had := os.LookupEnv("OPENAI_API_KEY")
	require.NoError(t, os.Unsetenv("OPENAI_API_KEY"))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("OPENAI_API_KEY", prev)
		}
	})

	var client *OpenAIClient
	var err error
	client, err = CreateClient(nil, nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrMissingAPIKey)
	require.Nil(t, client)
}

func TestCreateClientApiKeyFromConfig(t *testing.T) {
	// No OPENAI_API_KEY in the environment; the key comes from config.yaml
	// (key "api_key") instead (#228).
	prev, had := os.LookupEnv("OPENAI_API_KEY")
	require.NoError(t, os.Unsetenv("OPENAI_API_KEY"))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("OPENAI_API_KEY", prev)
		}
	})

	v := viper.New()
	v.Set("api_key", "from-config")

	client, err := CreateClient(v, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientApiKeyEnvOverridesConfig(t *testing.T) {
	// Both the environment variable and config.yaml provide a key; the
	// environment variable must win (#228).
	t.Setenv("OPENAI_API_KEY", "from-env")

	v := viper.New()
	v.Set("api_key", "from-config")

	client, err := CreateClient(v, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientApiKeyEmptyEnvFallsBackToConfig(t *testing.T) {
	// An OPENAI_API_KEY explicitly set to the empty string must not shadow
	// a usable config.yaml value.
	t.Setenv("OPENAI_API_KEY", "")

	v := viper.New()
	v.Set("api_key", "from-config")

	client, err := CreateClient(v, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientBaseURLFromConfig(t *testing.T) {
	// No OPENAI_API_BASE in the environment; the base URL comes from
	// config.yaml (key "base_url") instead, and is still validated (#228).
	t.Setenv("OPENAI_API_KEY", "test")
	prev, had := os.LookupEnv("OPENAI_API_BASE")
	require.NoError(t, os.Unsetenv("OPENAI_API_BASE"))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("OPENAI_API_BASE", prev)
		}
	})

	v := viper.New()
	v.Set("base_url", "https://api.openai.com/v1")

	client, err := CreateClient(v, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientBaseURLFromConfigRejectsUnsafeHost(t *testing.T) {
	// A public http:// host from config.yaml goes through the same SSRF
	// guard as the environment variable channel (#228).
	t.Setenv("OPENAI_API_KEY", "test")
	prev, had := os.LookupEnv("OPENAI_API_BASE")
	require.NoError(t, os.Unsetenv("OPENAI_API_BASE"))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("OPENAI_API_BASE", prev)
		}
	})

	v := viper.New()
	v.Set("base_url", "http://1.2.3.4/v1")

	client, err := CreateClient(v, nil)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "OPENAI_API_BASE")
}

func TestCreateClientBaseURLEnvOverridesConfig(t *testing.T) {
	// Both the environment variable and config.yaml provide a base URL; the
	// environment variable must win (#228). The config value would fail
	// validation if it were the one applied, so a successful client proves
	// the env value ("https://api.openai.com/v1") was used instead.
	t.Setenv("OPENAI_API_KEY", "test")
	t.Setenv("OPENAI_API_BASE", "https://api.openai.com/v1")

	v := viper.New()
	v.Set("base_url", "http://1.2.3.4/v1")

	client, err := CreateClient(v, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestCreateClientAPIBaseValidation(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		wantErr bool
	}{
		// https is always allowed
		{"openai", "https://api.openai.com/v1", false},
		{"openrouter", "https://openrouter.ai/api/v1", false},
		{"azure", "https://my-resource.openai.azure.com/openai/deployments/x", false},
		{"localhost over https", "https://localhost:8443/v1", false},

		// http allowed for loopback and private ranges (local LLM containers)
		{"http localhost", "http://localhost:8080/v1", false},
		{"http loopback ipv4", "http://127.0.0.1:11434/v1", false},
		{"http loopback ipv6", "http://[::1]:8080/v1", false},
		{"http rfc1918 10", "http://10.0.0.5/v1", false},
		{"http rfc1918 172.16", "http://172.16.0.1/v1", false},
		{"http rfc1918 192.168", "http://192.168.1.10:8080/v1", false},
		{"http ipv6 ula", "http://[fd00::1]/v1", false},

		// http rejected for public hosts and IMDS
		{"http openai", "http://api.openai.com/v1", true},
		{"http public ipv4", "http://1.2.3.4/v1", true},
		{"imds via http", "http://169.254.169.254/latest/meta-data/", true},
		{"http single-label hostname", "http://thinkbox:8080/v1", true},

		// explicit link-local / unspecified rejections
		{"http ipv6 link-local", "http://[fe80::1]/v1", true},
		{"http ipv4 unspecified", "http://0.0.0.0:11434/v1", true},
		{"http ipv4 zero-prefix", "http://0.1.2.3/v1", true},
		{"http cgnat", "http://100.64.0.1/v1", true},
		// IPv4-mapped IPv6 of IMDS — Go's IsLinkLocalUnicast classifies it correctly
		{"http ipv4-mapped imds", "http://[::ffff:169.254.169.254]/", true},
		// userinfo trick: localhost in userinfo, evil.com as host
		{"http userinfo trick", "http://localhost@evil.com/", true},

		// malformed / non-network schemes
		{"empty", "", true},
		{"missing scheme", "api.openai.com/v1", true},
		{"file scheme", "file:///etc/passwd", true},
		{"javascript scheme", "javascript:alert(1)", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("OPENAI_API_KEY", "test")
			t.Setenv("OPENAI_API_BASE", tc.baseURL)

			client, err := CreateClient(nil, nil)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, client)
				require.Contains(t, err.Error(), "OPENAI_API_BASE")
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
			}
		})
	}
}

func TestCreateClientAPIBaseInsecureOptOut(t *testing.T) {
	// insecureAPIBase=true bypasses validation entirely so single-label LAN
	// hostnames work (the reporter's http://thinkbox:8080/v1 case in #371).
	t.Setenv("OPENAI_API_KEY", "test")
	t.Setenv("OPENAI_API_BASE", "http://thinkbox:8080/v1")

	v := viper.New()
	v.Set("insecureAPIBase", true)

	client, err := CreateClient(v, nil)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestSimplePrompt(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"Say: Hello World!"}
	expected := "Hello World!"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "", prompt, "txt", nil)
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	// Cache dir should be empty
	cacheDir := testCtx.Config.GetString("cacheDir")
	err = filepath.Walk(cacheDir, func(path string, _ os.FileInfo, err error) error {
		if path == cacheDir {
			// Skip the root dir
			return nil
		}
		require.NoError(t, err)
		require.Fail(t, "Cache dir should be empty")
		return nil
	})
	require.NoError(t, err)

	wg.Wait()
}

func TestStreamSimplePrompt(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"Say: Hello World!"}
	expected := "Hello World!"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponseStream(expected)

	testCtx.Config.Set("stream", true)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "", prompt, "txt", nil)
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestPromptSaveAsChat(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"Say: Hello World!"}
	expected := "Hello World!"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt", nil)
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 2)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[0].Role)
	require.Equal(t, prompt[0], messages[0].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[1].Role)
	require.Equal(t, expected, messages[1].Content)

	wg.Wait()
}

func TestPromptLoadChat(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"Repeat last message"}
	expected := "World!"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
	require.NoError(t, err)

	err = manager.SaveSession("test_chat", []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "Hello",
		},
		{
			Role:    openai.ChatMessageRoleAssistant,
			Content: "World!",
		},
	})
	require.NoError(t, err)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt", nil)
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 4)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[2].Role)
	require.Equal(t, prompt[0], messages[2].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[3].Role)
	require.Equal(t, expected, messages[3].Content)

	wg.Wait()
}

func TestPromptWithModifier(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"Print Hello World!"}
	response := `echo \"Hello World\"`
	expected := `echo "Hello World"`

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(response)

	t.Setenv("SHELL", "/bin/bash")

	testCtx.Config.Set("chat", "test_chat")

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "sh", nil)
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 3)

	// Check if the modifier message was added
	require.Equal(t, openai.ChatMessageRoleSystem, messages[0].Role)

	// Check if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[1].Role)
	require.Equal(t, prompt[0], messages[1].Content)

	// Check if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[2].Role)
	require.Equal(t, expected, messages[2].Content)

	wg.Wait()
}

func TestSimplePromptWithLocalImage(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"what can you see on the picture?"}
	expected := "The image shows a character that appears to be a stylized robot. It has"
	inputImage := "testdata/marvin.jpg"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "", prompt, "txt", []string{inputImage})
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	wg.Wait()
}

func TestSimplePromptWithLocalImageAndChat(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"what can you see on the picture?"}
	expected := "The image shows a character that appears to be a stylized robot. It has"
	inputImage := "testdata/marvin.jpg"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt", []string{inputImage})
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 2)

	// Check, if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[0].Role)
	// The prompt should be empty, because it is a multi content message
	require.Empty(t, messages[0].Content)
	require.Len(t, messages[0].MultiContent, 2)
	// Check, if the prompt is a multi content message
	require.Equal(t, "text", string(messages[0].MultiContent[0].Type))
	require.Equal(t, prompt[0], messages[0].MultiContent[0].Text)
	// Check, if the image was added
	require.Equal(t, "image_url", string(messages[0].MultiContent[1].Type))
	require.NotEmpty(t, messages[0].MultiContent[1].ImageURL.URL)
	require.True(t, strings.HasPrefix(messages[0].MultiContent[1].ImageURL.URL, "data:"))

	// Check, if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[1].Role)
	require.Equal(t, expected, messages[1].Content)

	wg.Wait()
}

func TestSimplePromptWithURLImageAndChat(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"what can you see on the picture?"}
	expected := "The image shows a character that appears to be a stylized robot. It has"
	inputImage := "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt", []string{inputImage})
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 2)

	// Check, if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[0].Role)
	// The prompt should be empty, because it is a multi content message
	require.Empty(t, messages[0].Content)
	require.Len(t, messages[0].MultiContent, 2)
	// Check, if the prompt is a multi content message
	require.Equal(t, "text", string(messages[0].MultiContent[0].Type))
	require.Equal(t, prompt[0], messages[0].MultiContent[0].Text)
	// Check, if the image was added
	require.Equal(t, "image_url", string(messages[0].MultiContent[1].Type))
	require.Equal(t, inputImage, messages[0].MultiContent[1].ImageURL.URL)

	// Check, if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[1].Role)
	require.Equal(t, expected, messages[1].Content)

	wg.Wait()
}

func TestSimplePromptWithHTTPURLImageAndChat(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"what can you see on the picture?"}
	expected := "The image shows a character that appears to be a stylized robot. It has"
	inputImage := "http://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt", []string{inputImage})
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 2)

	// Check, if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[0].Role)
	// The prompt should be empty, because it is a multi content message
	require.Empty(t, messages[0].Content)
	require.Len(t, messages[0].MultiContent, 2)
	// Check, if the prompt is a multi content message
	require.Equal(t, "text", string(messages[0].MultiContent[0].Type))
	require.Equal(t, prompt[0], messages[0].MultiContent[0].Text)
	// Check, if the image was added as a URL (not treated as a local file)
	require.Equal(t, "image_url", string(messages[0].MultiContent[1].Type))
	require.Equal(t, inputImage, messages[0].MultiContent[1].ImageURL.URL)

	// Check, if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[1].Role)
	require.Equal(t, expected, messages[1].Content)

	wg.Wait()
}

func TestSimplePromptWithMixedImagesAndChat(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"what is the difference between those two pictures?"}
	expected := "The two images provided appear to be identical. Both show the same depiction of a"
	inputImageFile := "testdata/marvin.jpg"
	inputImageURL := "https://upload.wikimedia.org/wikipedia/en/c/cb/Marvin_%28HHGG%29.jpg"

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterExpectedChatResponse(expected)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, expected+"\n", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "test_chat", prompt, "txt", []string{inputImageURL, inputImageFile})
	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, writer.Close())

	require.FileExists(t, filepath.Join(testCtx.Config.GetString("cacheDir"), "test_chat"))

	var manager chat.SessionManager
	manager, err = chat.NewFilesystemChatSessionManager(testCtx.Config)
	require.NoError(t, err)

	var messages []openai.ChatCompletionMessage
	messages, err = manager.GetSession("test_chat")
	require.NoError(t, err)
	require.Len(t, messages, 2)

	// Check, if the prompt was added
	require.Equal(t, openai.ChatMessageRoleUser, messages[0].Role)
	// The prompt should be empty, because it is a multi content message
	require.Empty(t, messages[0].Content)
	require.Len(t, messages[0].MultiContent, 3)

	// Check, if the prompt is a multi content message
	require.Equal(t, "text", string(messages[0].MultiContent[0].Type))
	require.Equal(t, prompt[0], messages[0].MultiContent[0].Text)
	// Check, if the URL image was added
	require.Equal(t, "image_url", string(messages[0].MultiContent[1].Type))
	require.Equal(t, inputImageURL, messages[0].MultiContent[1].ImageURL.URL)
	// Check, if the file image was added
	require.Equal(t, "image_url", string(messages[0].MultiContent[2].Type))
	require.NotEmpty(t, messages[0].MultiContent[2].ImageURL.URL)
	require.True(t, strings.HasPrefix(messages[0].MultiContent[2].ImageURL.URL, "data:"))

	// Check, if the response was added
	require.Equal(t, openai.ChatMessageRoleAssistant, messages[1].Role)
	require.Equal(t, expected, messages[1].Content)

	wg.Wait()
}

func TestSimplePrompt_EmptyChoices(t *testing.T) {
	testCtx := testlib.NewTestCtx(t)
	testlib.SetAPIKey(t)
	testlib.SetAPIBase(t)

	var wg sync.WaitGroup
	reader, writer := io.Pipe()

	client, err := CreateClient(testCtx.Config, writer)
	require.NoError(t, err)

	prompt := []string{"Say: Hello World!"}

	httpmock.ActivateNonDefault(client.HTTPClient)
	t.Cleanup(httpmock.DeactivateAndReset)
	testlib.RegisterEmptyChatResponse()

	wg.Add(1)
	go func() {
		defer wg.Done()
		var buf bytes.Buffer
		_, errReader := io.Copy(&buf, reader)
		require.NoError(t, errReader)
		require.NoError(t, reader.Close())
		require.Equal(t, "", buf.String())
	}()

	var result string
	result, err = client.CreateCompletion(context.Background(), "", prompt, "txt", nil)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrEmptyResponse)
	require.Equal(t, "", result)
	require.NoError(t, writer.Close())

	wg.Wait()
}
