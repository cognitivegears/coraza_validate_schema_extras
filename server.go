package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/corazawaf/coraza/v3"
	"github.com/corazawaf/coraza/v3/types"
)

// Define a custom handler that includes the WAF instance
type validationHandler struct {
	waf coraza.WAF
}

// Implement the http.Handler interface for validationHandler
func (h *validationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Only POST requests are allowed\n")
		return
	}

	// Determine the preferred response format based on Accept header
	// Default to the same format as the request if Accept header is not specified
	requestContentType := r.Header.Get("Content-Type")
	acceptHeader := r.Header.Get("Accept")
	wantsJSON := (strings.Contains(acceptHeader, "application/json") ||
		strings.Contains(requestContentType, "application/json")) &&
		!strings.Contains(acceptHeader, "text/html")

	// Create a transaction
	tx := h.waf.NewTransaction()
	defer tx.ProcessLogging()

	// Process the request with Coraza manually
	// Parse client and host IP/port (default to 0 for ports if not present)
	clientAddr := r.RemoteAddr
	clientIP := clientAddr
	clientPort := 0
	if strings.Contains(clientAddr, ":") {
		parts := strings.Split(clientAddr, ":")
		clientIP = parts[0]
		if len(parts) > 1 {
			if port, err := strconv.Atoi(parts[1]); err == nil {
				clientPort = port
			}
		}
	}
	// Default server port to 80 if not specified
	serverPort := 80
	if r.TLS != nil {
		serverPort = 443
	}

	tx.ProcessConnection(clientIP, clientPort, r.Host, serverPort)
	tx.ProcessURI(r.URL.String(), r.Method, r.Proto)
	for name, values := range r.Header {
		for _, value := range values {
			tx.AddRequestHeader(name, value)
		}
	}
	tx.ProcessRequestHeaders()

	// Read the body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %v", err)
		if wantsJSON {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": "error", "message": "Failed to read request body", "code": "internal_error"}`)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal server error: Failed to read request body\n")
		}
		return
	}

	// Process the body
	_, _, err = tx.WriteRequestBody(bodyBytes)
	if err != nil {
		log.Printf("[ERROR] Failed to write request body: %v", err)
		if wantsJSON {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"status": "error", "message": "Failed to process request body", "code": "internal_error"}`)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Internal server error: Failed to process request body\n")
		}
		return
	}
	tx.ProcessRequestBody()

	// Check if the request was blocked
	if tx.IsInterrupted() {
		log.Printf("[CORAZA BLOCKED] Request with content-type %s blocked by security policy", requestContentType)
		if wantsJSON {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, `{"status": "error", "message": "Request blocked by security policy", "code": "schema_validation_failed"}`)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Error: Request blocked by security policy. Schema validation failed.\n")
		}
		return
	}

	// If we get here, the request was not blocked, so it passed validation
	log.Printf("[SUCCESS] Request with content-type %s passed validation", requestContentType)
	if wantsJSON {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status": "success", "message": "Validation successful! Your request passed all schema validations."}`)
	} else {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Validation successful! Your request passed all schema validations.\n")
	}
}

func main() {
	// Get default base directory (current working directory)
	defaultBaseDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Check for environment variable first, then fall back to default
	envRulesDir := os.Getenv("CORAZA_RULES_DIR")
	if envRulesDir == "" {
		envRulesDir = defaultBaseDir
	}

	// Parse flags
	port := flag.Int("port", 8080, "Port to listen on")
	rulesDirFlag := flag.String("rulesDir", envRulesDir, "Directory containing Coraza rules and schemas. Defaults to CORAZA_RULES_DIR env var or current directory.")
	flag.Parse()

	// Determine the effective rules directory (flag takes precedence over env var)
	rulesDir := *rulesDirFlag
	schemasDir := filepath.Join(rulesDir, "schemas") // Assume schemas are in a subdir of rulesDir

	log.Printf("Using rules directory: %s", rulesDir)
	log.Printf("Using schemas directory: %s", schemasDir) // Log schemas dir too

	// Construct the main rules file path
	mainRulesFile := filepath.Join(rulesDir, "rules", "main.conf") // Expect main.conf inside rules/ subdir

	// Check if the main rules file exists
	if _, err := os.Stat(mainRulesFile); os.IsNotExist(err) {
		log.Fatalf("Main rules file not found at %s. Ensure the rules directory structure is correct.", mainRulesFile)
	}

	// Initialize Coraza WAF using the specified rules directory
	wafConfig := coraza.NewWAFConfig().
		WithRootFS(os.DirFS(rulesDir)).                                                 // Set the root filesystem for Include directives
		WithDirectives(fmt.Sprintf("Include %s", filepath.Join("rules", "main.conf"))). // Relative path from rulesDir
		WithErrorCallback(func(rule types.MatchedRule) {
			log.Printf("[CORAZA ERROR] %s", rule.ErrorLog()) // Updated to match current API
		})

	waf, err := coraza.NewWAF(wafConfig)
	if err != nil {
		log.Fatalf("Failed to create WAF: %v", err)
	}

	log.Printf("Coraza WAF initialized with rules from %s", mainRulesFile)

	// Create the custom handler with the WAF instance
	validationH := &validationHandler{waf: waf}

	// Create mux
	mux := http.NewServeMux()
	mux.Handle("/validate", validationH) // Use the custom handler

	// Add route for static files (schemas) relative to the schemas directory
	mux.HandleFunc("/schemas/", func(w http.ResponseWriter, r *http.Request) {
		// Construct the full path relative to the schemasDir
		relPath := strings.TrimPrefix(r.URL.Path, "/schemas/")
		filePath := filepath.Join(schemasDir, relPath)

		// Basic security check to prevent path traversal
		if !strings.HasPrefix(filePath, schemasDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		log.Printf("Serving schema file: %s", filePath)
		http.ServeFile(w, r, filePath)
	})

	// Add informational endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
Schema Validation Test Server (Docker Ready)

Available Endpoints:
- POST /validate (with JSON content)
- GET /schemas/user.json (Path relative to configured schemas directory)

To test validation (assuming default port 8080):
1. For JSON: curl -X POST -H "Content-Type: application/json" --data @valid/valid_user.json http://localhost:%d/validate

Server listening on port %d. Rules loaded from: %s
`, *port, *port, mainRulesFile)
	})

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
