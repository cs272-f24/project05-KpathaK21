
package main

import(
	"context"
	"encoding/json"
	"strconv"
	"os"
	"log"
	"fmt"
	
	chroma "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
)

// Add adds a list of Course objects to the ChromaDB collection
func Add(courses []Course) (context.Context, *chroma.Client, *chroma.Collection) {
    openaikey := os.Getenv("OPENAI_PROJECT_KEY")
    if openaikey == "" {
        log.Fatalf("OPENAI_PROJECT_KEY not set in environment variables")
    }

    ctx := context.TODO()
    client, err := chroma.NewClient("http://localhost:8000")
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    openaiEf, err := openai.NewOpenAIEmbeddingFunction(openaikey)
    if err != nil {
        log.Fatalf("Error creating OpenAI embedding function: %s", err)
    }

    collection, err := client.GetCollection(ctx, "courses-collection", openaiEf)
    if err != nil {
        log.Fatalf("Failed to get collection: %v", err)
    }

    // Check if there are existing documents by performing a test query
    testQueryResults, err := collection.Query(ctx, []string{"test"}, 1, nil, nil, nil)
    if err == nil && len(testQueryResults.Documents) > 0 {
        fmt.Println("Courses already loaded in ChromaDB, skipping addition.")
        return ctx, client, collection
    }

    instructors := InitializeInstructors()
    for i, course := range courses {
        fullName := course.InstructorFirstName + " " + course.InstructorLastName
        canonicalName := findCanonicalName(fullName, instructors)

        metadata := map[string]interface{}{
            "instructor_canonical_name": canonicalName,
        }
        
        jsonData, err := json.Marshal(course)
        if err != nil {
            log.Printf("Failed to marshal course to JSON: %v", err)
            continue
        }

        _, err = collection.Add(ctx, nil, []map[string]interface{}{metadata}, []string{string(jsonData)}, []string{strconv.Itoa(i)})
        if err != nil {
            log.Printf("Failed to add course to collection: %v", err)
        } else {
            fmt.Printf("Added Course %v with Instructor %s\n", course.Title, canonicalName)
        }
    }
    return ctx, client, collection
}

// Query searches the ChromaDB collection for a term and retrieves matching documents
func Query(ctx context.Context, client *chroma.Client, collection *chroma.Collection, term string) [][]string {
    terms := []string{term}
    fmt.Printf("Querying for term: %s\n", term)

    queryResults, err := collection.Query(ctx, terms, 5, nil, nil, nil)
    if err != nil {
        log.Fatalf("failed to query collection: %v", err)
    }

    documents := queryResults.Documents

    return documents
}
