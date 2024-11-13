package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	
)

func RealChatBot() *ChatBot {
	apiKey := os.Getenv("OPENAI_PROJECT_KEY")
	if apiKey == "" {
		log.Fatal("API key is missing. Please set OPENAI_PROJECT_KEY environment variable.")
	}
	llmClient := NewLLMClient(apiKey)

	csvFilePath := "Fall 2024 Class Schedule 08082024.csv"
	csvFile, err := os.Open(csvFilePath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer csvFile.Close()

	courses, err := ReadCSV(csvFile)
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	fmt.Printf("Loaded %d courses from CSV:\n", len(courses))
	metadataExtractor := &MetadataExtractor{courses: courses}

	chromaCtx, chromaClient, collection := Add(courses)
	return NewChatBot(llmClient, metadataExtractor, chromaCtx, chromaClient, collection)
}

// TestPHIL tests the ChatBot response for philosophy courses
func TestPHIL(t *testing.T) {
	chatbot := RealChatBot()
	question := "Which philosophy courses are offered this semester?"
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil || !strings.Contains(answer, "Great Philosophical Questions") {
		t.Errorf("Expected answer to mention 'Great Philosophical Questions', got: %v", answer)
	}
}

func TestPhil(t *testing.T) {
	chatbot := RealChatBot()
	question := "What courses is Phil Peterson teaching in Fall 2024?"

	// Expected instructor to appear in the answer
	expectedInstructor := "Philip Peterson"

	// Get the answer for the question
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the answer contains the expected instructor name
	if !strings.Contains(answer, expectedInstructor) {
		t.Errorf("Expected answer to contain instructor '%s', got:\n%v", expectedInstructor, answer)
	}
}

// TestBio tests the ChatBot response for Bioinformatics location
func TestBio(t *testing.T) {
	chatbot := RealChatBot()
	question := "Where does Bioinformatics meet?"
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil || !strings.Contains(answer, "KA") || !strings.Contains(answer, "311") {
		t.Errorf("Expected answer to mention 'Building KA, Room 311', got: %v", answer)
	}
}

func TestGuitar(t *testing.T) {
	chatbot := RealChatBot()
	question := "Can I learn guitar this semester?"
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil || !strings.Contains(answer, "Guitar and Bass Lessons") {
		t.Errorf("Expected answer to mention 'Guitar and Bass Lessons', got: %v", answer)
	}
}

func TestMultiple(t *testing.T) {
	chatbot := RealChatBot()
	question := "I would like to take a Rhetoric course from Phil Choong. What can I take?"
	answer, err := chatbot.AnswerQuestion(question)
	fmt.Printf("Answer for question '%s':\n%s\n", question, answer)
	if err != nil || !strings.Contains(answer, "Philip Choong") {
		t.Errorf("Expected answer to contain 'Philip Choong', got: %v", answer)
	}
}



// package main
// 
// import (
//     "testing"
//     "strings"
//     "os"
//     "log"
//     "fmt"
// 
// )
// 
// func RealChatBot() *ChatBot {
//     apiKey := os.Getenv("OPENAI_PROJECT_KEY")
//     if apiKey == "" {
//         log.Fatal("API key is missing. Please set OPENAI_API_KEY environment variable.")
//     }
//     llmClient := NewLLMClient(apiKey)
// 
//     csvFilePath := "Fall 2024 Class Schedule 08082024.csv"
//     csvFile, err := os.Open(csvFilePath)
//     if err != nil {
//         log.Fatalf("Failed to open CSV file: %v", err)
//     }
//     defer csvFile.Close()
// 
//     courses, err := ReadCSV(csvFile)
//     if err != nil {
//         log.Fatalf("Failed to read CSV file: %v", err)
//     }
// 
//     fmt.Printf("Loaded %d courses from CSV:\n", len(courses))
//     metadataExtractor := &MetadataExtractor{courses: courses}
// 
//     chromaCtx, chromaClient, collection := Add(courses)
//     return NewChatBot(llmClient, metadataExtractor, chromaCtx, chromaClient, collection)
// }
// 
// 
// //TestPHIL tests the ChatBot response for philosophy courses
// func TestPHIL(t *testing.T) {
// 	chatbot := RealChatBot()
//     question := "Which philosophy courses are offered this semester?"
//     answer, err := chatbot.AnswerQuestion(question)
//     if err != nil || !strings.Contains(answer, "Great Philosophical Questions") {
//         t.Errorf("Expected answer to mention 'Great Philosophical Questions', got: %v", answer)
//     }
// }
// 
// // TestPhil tests the ChatBot response for courses taught by Phil Peterson
// func TestPhil(t *testing.T) {
// 	chatbot := RealChatBot()
// 	question := "What courses is Phil Peterson teaching in Fall 2024?"
// 
// 	// Expected instructor to appear in the answer
// 	expectedInstructor := "Philip Peterson"
// 
// 	// Get the answer for the question
// 	answer, err := chatbot.AnswerQuestion(question)
// 	if err != nil {
// 		t.Fatalf("Expected no error, got %v", err)
// 	}
// 
// 	// Verify the answer contains the expected instructor name
// 	if !strings.Contains(answer, expectedInstructor) {
// 		t.Errorf("Expected answer to contain instructor '%s', got:\n%v", expectedInstructor, answer)
// 	}
// }
// 
// // TestBio tests the ChatBot response for Bioinformatics location
// func TestBio(t *testing.T) {
// 	chatbot := RealChatBot()
//     question := "Where does Bioinformatics meet?"
//     answer, err := chatbot.AnswerQuestion(question)
//     if err != nil || !strings.Contains(answer, "KA") || !strings.Contains(answer, "311") {
//         t.Errorf("Expected answer to mention 'Building KA, Room 311', got: %v", answer)
//     }
// }
// 
// func TestGuitar(t *testing.T) {
// 	chatbot := RealChatBot()
//     question := "Can I learn guitar this semester?"
//     answer, err := chatbot.AnswerQuestion(question)
//     if err != nil || !strings.Contains(answer, "Guitar and Bass Lessons") {
//         t.Errorf("Expected answer to mention 'Guitar and Bass Lessons', got: %v", answer)
//     }
// }
// 
// func TestMultiple(t *testing.T) {
// 	chatbot := RealChatBot()
//     question := "I would like to take a Rhetoric course from Phil Choong. What can I take?"
//     answer, err := chatbot.AnswerQuestion(question)
//     if err != nil || !strings.Contains(answer, "Philip Choong") {
//         t.Errorf("Expected answer to contain 'Philip Choong', got: %v", answer)
//     }
// }
