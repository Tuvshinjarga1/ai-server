package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

type MCPCallRequest struct {
	Function string                 `json:"function"`
	Args     map[string]interface{} `json:"args"`
}

type MCPCallResponse struct {
	Result interface{} `json:"result"`
}

func main() {
	// Get port from environment variable (Railway sets PORT)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up HTTP server
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/process", handleProcess)
	http.HandleFunc("/health", handleHealth)

	log.Printf("Starting server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "AI Server is running",
		"status":  "healthy",
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_users",
				Description: "Get users from the database",
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "create_absence_request",
				Description: "Create absence request",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"user_email": map[string]interface{}{
							"type":        "string",
							"description": "The email address of the user requesting absence",
						},
						"start_date": map[string]interface{}{
							"type":        "string",
							"description": "The start date of the absence (YYYY-MM-DD)",
						},
						"end_date": map[string]interface{}{
							"type":        "string",
							"description": "The end date of the absence (YYYY-MM-DD)",
						},
						"reason": map[string]interface{}{
							"type":        "string",
							"description": "The reason for the absence",
						},
						"in_active_hours": map[string]interface{}{
							"type":        "number",
							"description": "The number of hours the user will be inactive. If not provided, it will be calculated based on the start and end date. 1 day = 8 hour",
						},
					},
					"required": []string{"user_email", "start_date", "end_date", "reason", "in_active_hours"},
				},
			},
		},
	}

	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: "Please create absence request for user with email test_user10@fibo.cloud for 2 days starting from 2025-07-02. Reason: I'm sick"},
		},
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: 0.7,
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("OpenAI API error: %v", err), http.StatusInternalServerError)
		return
	}

	msg := resp.Choices[0].Message
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"message": msg.Content,
	}

	if len(msg.ToolCalls) > 0 {
		tool := msg.ToolCalls[0]
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(tool.Function.Arguments), &args); err != nil {
			http.Error(w, fmt.Sprintf("Error parsing tool arguments: %v", err), http.StatusInternalServerError)
			return
		}

		result := callMCP(tool.Function.Name, args)
		response["tool_call"] = map[string]interface{}{
			"function": tool.Function.Name,
			"args":     args,
			"result":   result,
		}
	}

	json.NewEncoder(w).Encode(response)
}

func callMCP(function string, args map[string]interface{}) interface{} {
	body, _ := json.Marshal(MCPCallRequest{
		Function: function,
		Args:     args,
	})
	
	// Get MCP server URL from environment variable, default to localhost
	mcpServerURL := os.Getenv("MCP_SERVER_URL")
	if mcpServerURL == "" {
		mcpServerURL = "http://localhost:8080"
	}
	
	resp, err := http.Post(mcpServerURL+"/call-function", "application/json", bytes.NewReader(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result MCPCallResponse
	json.Unmarshal(data, &result)
	return result.Result
}
