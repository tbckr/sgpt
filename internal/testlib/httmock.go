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

package testlib

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/jarcoal/httpmock"
)

const (
	baseURL              = "https://api.openai.com/v1"
	chatCompletionSuffix = "/chat/completions"
)

func RegisterExpectedChatResponse(response string) {
	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("%s%s", baseURL, chatCompletionSuffix),
		httpmock.NewStringResponder(
			200,
			fmt.Sprintf(`{
				"choices": [
					{
						"index": 0,
						"finish_reason": "length",
						"message": {
							"role": "assistant",
							"content": "%s"
						}
					}
				]
			}`, response),
		),
	)
}

func RegisterExpectedChatResponseStream(response string) {
	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("%s%s", baseURL, chatCompletionSuffix),
		func(request *http.Request) (*http.Response, error) {
			// Reference: https://github.com/sashabaranov/go-openai/blob/a09cb0c528c110a6955a9ee9a5d021a57ed44b90/chat_stream_test.go#L39
			data := createStreamedMessages(response)
			resp := httpmock.NewBytesResponse(200, data)
			resp.Header.Set("Content-Type", "text/event-stream")
			return resp, nil
		},
	)
}

func createStreamedMessages(response string) []byte {
	const (
		eventMessage = "event: message\n"

		dataMessageTemplate = "data: %s\n\n"
		messageTemplate     = `{"id":"1","object":"completion","created":1598069254,"model":"gpt-3.5-turbo","choices":[{"index":0,"delta":{"content":"%c"},"finish_reason":"max_tokens"}]}`

		eventDone = "event: done\n"
		dataDone  = "data: [DONE]\n\n"
	)
	var buff bytes.Buffer
	for _, char := range response {
		buff.WriteString(eventMessage)
		buff.WriteString(fmt.Sprintf(dataMessageTemplate, fmt.Sprintf(messageTemplate, char)))
	}
	buff.WriteString(eventDone)
	buff.WriteString(dataDone)

	return buff.Bytes()
}
