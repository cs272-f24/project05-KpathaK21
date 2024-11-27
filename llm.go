package main

import (
    "context"
    "fmt"
    "strings"

    openai "github.com/sashabaranov/go-openai"
)

// LLMClient is a reusable abstraction for interacting with the chat completion API.
// It encapsulates the OpenAI client to allow easy integration and interaction with OpenAI's LLM services.
type LLMClient struct {
    client *openai.Client // The underlying OpenAI client instance used to interact with the API.
}

// NewLLMClient initializes a new LLMClient with an API key.
// The API key is required to authenticate requests to the OpenAI API.
// Params:
// - apiKey: The OpenAI API key as a string.
// Returns:
// - An initialized LLMClient instance.
func NewLLMClient(apiKey string) *LLMClient {
    client := openai.NewClient(apiKey) // Create a new OpenAI client instance using the provided API key.
    return &LLMClient{client: client} // Return an LLMClient with the created OpenAI client.
}

// ChatCompletion takes a question, replaces aliases with canonical names, and returns the LLM's response.
// This function processes the question by replacing any recognized aliases with their canonical names
// to ensure clarity for the LLM. It then sends the processed question and system message to the chat API.
// Params:
// - question: The user's query as a string.
// - systemMessage: A system message to set context for the chat completion.
// Returns:
// - The LLM's response as a string.
// - An error if the API call fails or another issue occurs.
func (llm *LLMClient) ChatCompletion(question, systemMessage string) (string, error) {
    instructors := InitializeInstructors() // Initialize the list of instructors with aliases and canonical names.
    
    // Replace all instructor aliases in the question with their canonical names.
    for _, instructor := range instructors {
        for _, alias := range instructor.Aliases {
            // Perform a case-insensitive match and replacement.
            if strings.Contains(strings.ToLower(question), strings.ToLower(alias)) {
                question = strings.ReplaceAll(question, alias, instructor.CanonicalName)
            }
        }
    }

    // Create a chat completion request with the specified system message and question.
    req := openai.ChatCompletionRequest{
        Model: openai.GPT4oMini, // Specify the model to use for the completion.
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem, // The system message to guide the LLM.
                Content: systemMessage,
            },
            {
                Role:    openai.ChatMessageRoleUser, // The user's query after processing.
                Content: question,
            },
        },
    }

    // Call the OpenAI API to get a chat completion.
    resp, err := llm.client.CreateChatCompletion(context.Background(), req)
    if err != nil {
        // Return a wrapped error to provide more context about the failure.
        return "", fmt.Errorf("CreateChatCompletion failed: %w", err)
    }

    // Return the content of the LLM's response message.
    return resp.Choices[0].Message.Content, nil
}
