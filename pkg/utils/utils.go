package utils

import (
	"fmt"
	"log"
	"net/http"
)

// LogRequest is a utility function that logs HTTP requests.
func LogRequest(r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.Path)
}

// HandleError is a utility function that handles errors by logging them and sending a response.
func HandleError(w http.ResponseWriter, err error, statusCode int) {
	log.Printf("Error: %v", err)
	http.Error(w, err.Error(), statusCode)
}

// FormatResponse is a utility function that formats the response data.
func FormatResponse(data any) string {
	// Implement your formatting logic here``
	return fmt.Sprintf("%v", data)
}