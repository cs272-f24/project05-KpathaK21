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

// QueryCourses processes guitar-related queries and instructor-based queries
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
    instructorFound := ""

    // Identify if the question mentions an instructor alias and find the canonical name.
    for _, instructor := range instructors {
        for _, alias := range instructor.Aliases {
            if strings.Contains(strings.ToLower(question), strings.ToLower(alias)) {
                instructorFound = instructor.CanonicalName
                fmt.Printf("Instructor found in question: %s\n", instructorFound)
                break
            }
        }
        if instructorFound != "" {
            break
        }
    }

    // Query for courses if the instructor is found
    if instructorFound != "" {
        fmt.Printf("Querying courses for instructor '%s'\n", instructorFound)
        var instructorCourses []string

        if bot.metadata.courses == nil || len(bot.metadata.courses) == 0 {
            return "", fmt.Errorf("no courses data available")
        }

        // Debugging: Print each course's instructor name to ensure matching
        for _, course := range bot.metadata.courses {
            fullInstructorName := strings.TrimSpace(course.InstructorFirstName + " " + course.InstructorLastName)
            //fmt.Printf("Checking course '%s' with instructor '%s' (Expected: '%s')\n", course.Title, fullInstructorName, instructorFound)

            if strings.EqualFold(fullInstructorName, instructorFound) {
                courseInfo := fmt.Sprintf(
                    "SUBJ                      : %s\n"+
                        "CRSE NUM                  : %s\n"+
                        "SEC                       : %s\n"+
                        "CRN                       : %s\n"+
                        "Schedule Type Code        : %s\n"+
                        "Campus Code               : %s\n"+
                        "Title Short Desc          : %s\n"+
                        "Instruction Mode Desc     : %s\n"+
                        "Meeting Type Codes        : %s\n"+
                        "Meet Days                 : %s\n"+
                        "Begin Time                : %s\n"+
                        "End Time                  : %s\n"+
                        "Meet Start                : %s\n"+
                        "Meet End                  : %s\n"+
                        "BLDG                      : %s\n"+
                        "RM                        : %s\n"+
                        "Actual Enrollment         : %s\n"+
                        "Primary Instructor First Name : %s\n"+
                        "Primary Instructor Last Name : %s\n"+
                        "Primary Instructor Email  : %s\n"+
                        "College                   : %s\n"+
                        strings.Repeat("-", 50),
                    course.Subject, course.CourseNumber, course.Section, course.CRN,
                    course.ScheduleTypeCode, course.CampusCode, course.Title,
                    course.InstructionModeDesc, course.MeetingTypeCodes, course.MeetDays,
                    course.BeginTime, course.EndTime, course.MeetStart, course.MeetEnd,
                    course.Building, course.Room, course.ActualEnrollment,
                    course.InstructorFirstName, course.InstructorLastName, course.InstructorEmail,
                    course.College,
                )
                instructorCourses = append(instructorCourses, courseInfo)
                fmt.Printf("Appending course for %s: %s\n", instructorFound, courseInfo) // Debug line
            }
        }

        // Print courses if matches are found
        if len(instructorCourses) > 0 {
            result := fmt.Sprintf("Here's what I found for %s:\n%s", instructorFound, strings.Join(instructorCourses, "\n"))
            fmt.Printf("Final response:\n%s\n", result) // Debug line
            return result, nil
        }
        return fmt.Sprintf("No courses found for %s.", instructorFound), nil
    }

    // If no instructor is found, perform a general search
    result := bot.QueryCourses(question)
    if len(result) > 0 {
        return fmt.Sprintf("Here's what I found:\n%v", result), nil
    }

    // Default system message if no match is found
    systemMessage := bot.generateSystemMessage(question)
    return bot.llmClient.ChatCompletion(question, systemMessage)
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
