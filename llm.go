package main

import (
    "context"
    "fmt"
    "strings"

    openai "github.com/sashabaranov/go-openai"
)

// LLMClient is a reusable abstraction for interacting with the chat completion API
type LLMClient struct {
    client *openai.Client
 
}

// NewLLMClient initializes a new LLMClient with an API key
func NewLLMClient(apiKey string) *LLMClient {
    client := openai.NewClient(apiKey)
    return &LLMClient{client: client}
}

// ChatCompletion takes a question, replaces aliases with canonical names, and returns the LLM's response
func (llm *LLMClient) ChatCompletion(question, systemMessage string) (string, error) {
    instructors := InitializeInstructors()
    for _, instructor := range instructors {
        for _, alias := range instructor.Aliases {
            if strings.Contains(strings.ToLower(question), strings.ToLower(alias)) {
                question = strings.ReplaceAll(question, alias, instructor.CanonicalName)
            }
        }
    }

    req := openai.ChatCompletionRequest{
        Model: openai.GPT4oMini,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem,
                Content: systemMessage,
            },
            {
                Role:    openai.ChatMessageRoleUser,
                Content: question,
            },
        },
    }

    resp, err := llm.client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        return "", fmt.Errorf("CreateChatCompletion failed: %w", err)
    }
    return resp.Choices[0].Message.Content, nil
}
