package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
)

var chatbot *ChatBot

func main() {
    apiKey := os.Getenv("OPENAI_PROJECT_KEY")
    csvFilePath := "Fall 2024 Class Schedule 08082024.csv"

    llmClient := NewLLMClient(apiKey)

    // Initialize metadata extractor only to check if the data needs to be loaded
    metadataExtractor, err := NewMetadataExtractor(csvFilePath, llmClient)
    if err != nil {
        log.Fatalf("Failed to initialize MetadataExtractor: %v", err)
    }

    // Set up chatbot with already-populated data in ChromaDB
    chromaCtx, chromaClient, collection := Add(metadataExtractor.courses)
    chatbot = NewChatBot(llmClient, metadataExtractor, chromaCtx, chromaClient, collection)

    fmt.Println("Entering interactive mode. Type your questions below:")
    runInteractiveMode()
}

// runInteractiveMode handles user queries interactively.
func runInteractiveMode() {
    scanner := bufio.NewScanner(os.Stdin)
    fmt.Print("\nCatalog search> ")
    for scanner.Scan() {
        question := scanner.Text()
        answer, err := chatbot.AnswerQuestion(question)
        if err != nil {
            fmt.Println("Error:", err)
            continue
        }
        fmt.Println(answer)
        fmt.Print("\nCatalog search> ")
    }
    if err := scanner.Err(); err != nil {
        log.Println("Error reading input:", err)
    }
}
