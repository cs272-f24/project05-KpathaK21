package main

import (
    "context"
    "fmt"
    "strings"

    chroma "github.com/amikos-tech/chroma-go"
)

// Instructor represents an instructor with a canonical name and aliases
type Instructor struct {
    CanonicalName string
    Aliases       []string
}

// InitializeInstructors creates a list of instructors with canonical names
func InitializeInstructors() []Instructor {
    return []Instructor{
        {CanonicalName: "Philip Peterson", Aliases: []string{"Phil Peterson", "Philip Peterson"}},
        {CanonicalName: "Philip Choong", Aliases: []string{"Phil Choong", "Philip Choong"}},
    }
}

// ChatBot uses LLMClient and MetadataExtractor to answer questions
type ChatBot struct {
    llmClient    *LLMClient
    metadata     *MetadataExtractor
    chromaCtx    context.Context
    chromaClient *chroma.Client
    collection   *chroma.Collection
}

// NewChatBot initializes a ChatBot with an LLM client, metadata extractor, and ChromaDB context
func NewChatBot(llmClient *LLMClient, metadata *MetadataExtractor, chromaCtx context.Context, chromaClient *chroma.Client, collection *chroma.Collection) *ChatBot {
    return &ChatBot{
        llmClient:    llmClient,
        metadata:     metadata,
        chromaCtx:    chromaCtx,
        chromaClient: chromaClient,
        collection:   collection,
    }
}

func findCanonicalName(inputName string, instructors []Instructor) string {
    inputName = strings.TrimSpace(inputName) // Remove leading and trailing whitespace
    if inputName == "" {
        return inputName // Return immediately if the input is empty
    }

    for _, instructor := range instructors {
        for _, alias := range instructor.Aliases {
            if strings.EqualFold(inputName, alias) {
                fmt.Printf("Substituting alias '%s' with canonical name '%s'\n", inputName, instructor.CanonicalName)
                return instructor.CanonicalName
            }
        }
    }
    fmt.Printf("No canonical substitution found for '%s'. Using original name.\n", inputName)
    return inputName // Return the original name if no match is found
}

// PrettyPrintDocuments formats and prints course documents nicely
func PrettyPrintDocuments(documents [][]string) {
    for i, doc := range documents {
        fmt.Printf("Course %d:\n", i+1) // Label each course entry

        docContent := strings.Join(doc, " ")
        fields := strings.Split(docContent, ". ")

        for _, field := range fields {
            if strings.TrimSpace(field) != "" {
                parts := strings.SplitN(field, ":", 2)
                if len(parts) == 2 {
                    fmt.Printf("%-25s  %s\n", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
                } else {
                    fmt.Println(field)
                }
            }
        }

        if i < len(documents)-1 {
            fmt.Println(strings.Repeat("-", 50)) // Line between courses
        }
        fmt.Println() // Blank line for readability
    }
}

func (bot *ChatBot) QueryCourses(term string) string {

    // Initialize the list of instructors and find the canonical name
    instructors := InitializeInstructors()
    canonicalName := findCanonicalName(term, instructors)
    fmt.Printf("Canonical name to search for: %s\n", canonicalName) // Debug line

    // Check if the query term matches an instructor alias
    if canonicalName != term {
        fmt.Printf("Searching for courses taught by: %s\n", canonicalName)
        
        // Declare instructorCourses as a slice of strings to store matched courses
        var instructorCourses []string
        
        // Iterate over courses to find matches for the instructor
        for _, course := range bot.metadata.courses {
            // Build the full instructor name from the course data
            fullInstructorName := strings.TrimSpace(course.InstructorFirstName + " " + course.InstructorLastName)
            fmt.Printf("Checking course: %s taught by %s\n", course.Title, fullInstructorName) // Debug line

            // Match based on the full name of the instructor
            if strings.EqualFold(fullInstructorName, canonicalName) {
                // Format the course information and append it to instructorCourses
                courseInfo := fmt.Sprintf("%s, Section: %s, CRN: %s in %s, Room %s",
                    course.Title, course.Section, course.CRN, course.Building, course.Room)
                instructorCourses = append(instructorCourses, courseInfo)
                fmt.Printf("Appending course for %s: %s\n", canonicalName, courseInfo) // Debug line
            }
        }

        // Check if any courses were found for the instructor and return the result
        if len(instructorCourses) > 0 {
            result := fmt.Sprintf("Here are the courses taught by %s:\n%s", canonicalName, strings.Join(instructorCourses, "\n"))
            fmt.Printf("Courses found for %s:\n%s\n", canonicalName, result) // Debug line
            return result
        }
        // Return a message if no courses were found for the instructor
        return fmt.Sprintf("No courses found for %s.", canonicalName)
    }

    // General query fallback if no instructor match
    results := Query(bot.chromaCtx, bot.chromaClient, bot.collection, term)
    PrettyPrintDocuments(results)
    if len(results) == 0 {
        return "No relevant courses found."
    }

    // Build the output for general query results
    var output strings.Builder
    output.WriteString(bot.metadata.header + "\n")

    for _, doc := range results {
        output.WriteString(strings.Join(doc, ",") + "\n")
    }

    return output.String()
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

    // Use the Query function to search the ChromaDB collection for relevant data
    documents := Query(bot.chromaCtx, bot.chromaClient, bot.collection, question)

    if len(documents) > 0 {
        preamble := "Based on the available course information, here are the relevant matches:\n\n"
        for _, doc := range documents {
            preamble += fmt.Sprintf("- %s\n", strings.Join(doc, " "))
        }
        preamble += "\nPlease use this information to answer the user's question."

        // Use ChatCompletion with the preamble
        response, err := bot.llmClient.ChatCompletion(question, preamble)
        if err != nil {
            return "", fmt.Errorf("ChatCompletion failed: %w", err)
        }
        return response, nil
    }

    // Fallback: general system message
    systemMessage := "Provide accurate information based on the context of university courses."
    response, err := bot.llmClient.ChatCompletion(question, systemMessage)
    if err != nil {
        return "", fmt.Errorf("ChatCompletion failed: %w", err)
    }
    return response, nil
}

// needsDirectQuery determines if the question should be answered by querying ChromaDB directly
func (bot *ChatBot) needsDirectQuery(question string) bool {
    questionLower := strings.ToLower(question)
    return containsAny(questionLower, bot.metadata.Instructors) || 
           containsAny(questionLower, bot.metadata.Departments) || 
           strings.Contains(questionLower, "location") || 
           strings.Contains(questionLower, "meeting")
}

// generateSystemMessage constructs a relevant prompt based on question context
func (bot *ChatBot) generateSystemMessage(question string) string {
    questionLower := strings.ToLower(question)

    if containsAny(questionLower, bot.metadata.Instructors) {
        return "Answer questions related to instructors, including their courses and schedules."
    } else if containsAny(questionLower, bot.metadata.Departments) {
        return "Answer questions about available courses by department."
    } else if strings.Contains(questionLower, "location") || strings.Contains(questionLower, "where") {
        return "Provide location and timing information for requested courses."
    }
    return "Answer questions about USF courses by department, course name, instructor, times, and location."
}

// containsAny checks if a question contains any of the given values (keywords).
func containsAny(question string, values []string) bool {
    for _, value := range values {
        if strings.Contains(strings.ToLower(question), strings.ToLower(value)) {
            return true
        }
    }
    return false
}
