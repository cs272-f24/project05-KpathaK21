package main

import (
    "context"
    "fmt"
    "strings"
	"log"
    chroma "github.com/amikos-tech/chroma-go"
)



// ChatBot uses LLMClient and MetadataExtractor to answer questions
type ChatBot struct {
    llmClient            *LLMClient
    metadata             *MetadataExtractor
    chromaCtx            context.Context
    chromaClient         *chroma.Client
    courseCollection     *chroma.Collection
    instructorCollection *chroma.Collection
}


// NewChatBot initializes a ChatBot with an LLM client, metadata extractor, and ChromaDB context
func NewChatBot(llmClient *LLMClient, metadata *MetadataExtractor, chromaCtx context.Context, chromaClient *chroma.Client, courseCollection, instructorCollection *chroma.Collection) *ChatBot {
    return &ChatBot{
        llmClient:         llmClient,
        metadata:          metadata,
        chromaCtx:         chromaCtx,
        chromaClient:      chromaClient,
        courseCollection:  courseCollection,
        instructorCollection: instructorCollection,
    }
}




func (bot *ChatBot) QueryCourses(term string) string {
    // Find the canonical name for the given term
    instructors := InitializeInstructors()
    canonicalName := findCanonicalName(term, instructors)
    fmt.Printf("Canonical name to search for: %s\n", canonicalName)

    // If the canonical name is empty, return a fallback message
    if canonicalName == "" {
        return fmt.Sprintf("No valid instructor found for '%s'.", term)
    }

	
	
    // Query the collection using the canonical name
    queryResults, err := bot.courseCollection.Query(bot.chromaCtx, []string{canonicalName}, 5, nil, nil, nil)
    if err != nil {
        log.Printf("Error querying collection: %v", err)
        return "An error occurred while searching for courses."
    }

    // Check if results are empty
    if len(queryResults.Documents) == 0 {
        return fmt.Sprintf("No courses found for %s.", canonicalName)
    }

    // Format the results
    var result strings.Builder
    result.WriteString(fmt.Sprintf("Here are the courses taught by %s:\n", canonicalName))
    for _, doc := range queryResults.Documents {
        result.WriteString(fmt.Sprintf("- %s\n", doc))
    }

    return result.String()
}


func (bot *ChatBot) AnswerQuestion(question string) (string, error) {
    fmt.Printf("Processing question: %s\n", question)

    instructors := InitializeInstructors()
    for _, instructor := range instructors {
        for _, alias := range instructor.Aliases {
            if strings.Contains(strings.ToLower(question), strings.ToLower(alias)) {
                question = strings.ReplaceAll(question, alias, instructor.CanonicalName)
            }
        }
    }

    // Use the appropriate collection for the query
    var collectionToQuery *chroma.Collection
    if strings.Contains(strings.ToLower(question), "instructor") {
        collectionToQuery = bot.instructorCollection
    } else {
        collectionToQuery = bot.courseCollection
    }

    documents := Query(bot.chromaCtx, bot.chromaClient, collectionToQuery, question)

    if len(documents) > 0 {
        preamble := "Based on the available information, here are the relevant matches:\n\n"
        for _, doc := range documents {
            preamble += fmt.Sprintf("- %s\n", strings.Join(doc, " "))
        }
        preamble += "\nPlease use this information to answer the user's question."

        response, err := bot.llmClient.ChatCompletion(question, preamble)
        if err != nil {
            return "", fmt.Errorf("ChatCompletion failed: %w", err)
        }
        return response, nil
    }

    // Fallback: general system message
    systemMessage := "Provide accurate information based on the context of university courses and instructors."
    response, err := bot.llmClient.ChatCompletion(question, systemMessage)
    if err != nil {
        return "", fmt.Errorf("ChatCompletion failed: %w", err)
    }
    return response, nil
}

