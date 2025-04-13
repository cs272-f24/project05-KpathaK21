package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"

    chroma "github.com/amikos-tech/chroma-go"
)

var chatbot *ChatBot

func main() {
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("API key is missing. Please set OPENAI_API_KEY environment variable.")
    }
    csvFilePath := "Fall 2024 Class Schedule 08082024.csv"

    llmClient := NewLLMClient(apiKey)

    // Initialize metadata extractor
    metadataExtractor, err := NewMetadataExtractor(csvFilePath, llmClient)
    if err != nil {
        log.Fatalf("Failed to initialize MetadataExtractor: %v", err)
    }

    // Add courses and instructors to ChromaDB
    chromaCtx, chromaClient, courseCollection, instructorCollection := Add(metadataExtractor.courses)

    // Initialize chatbot with collections
    chatbot = NewChatBot(llmClient, metadataExtractor, chromaCtx, chromaClient, courseCollection, instructorCollection)

    fmt.Println("Courses and instructors added to collections.")
    fmt.Println("Entering interactive mode. Type your questions below:")
    runInteractiveMode(chromaCtx, chromaClient, courseCollection)
}

// runInteractiveMode handles user queries interactively.
func runInteractiveMode(ctx context.Context, client *chroma.Client, collection *chroma.Collection) {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Print("\nCatalog search> ")
    for scanner.Scan() {
        question := scanner.Text()
        if question == "" {
            fmt.Println("Please enter a valid query.")
            continue
        }

        // Use the chatbot's AnswerQuestion method
        answer, err := chatbot.AnswerQuestion(question)
        if err != nil {
            fmt.Printf("Error processing your question: %v\n", err)
            continue
        }

        fmt.Println(answer) // Print the chatbot's answer
        fmt.Print("\nCatalog search> ")
    }

    if err := scanner.Err(); err != nil {
        log.Println("Error reading input:", err)
    }
}

