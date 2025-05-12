package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/corazawaf/coraza/v3"
	txhttp "github.com/corazawaf/coraza/v3/http"
	"github.com/corazawaf/coraza/v3/types"
)

// Define a custom handler that includes the WAF instance
type validationHandler struct {
	waf coraza.WAF
}

// Create a http.Handler for successful validation
func successHandler(w http.ResponseWriter, r *http.Request) {

	// Check Accept header for JSON or HTML
	acceptHeader := r.Header.Get("Accept")

	// Default to JSON if Accept header is not specified
	wantsJSON := strings.Contains(acceptHeader, "application/json")
	if wantsJSON {
		w.Header().Set("Content-Type", "application/json")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}

	w.WriteHeader(http.StatusOK)

	if wantsJSON {
		fmt.Fprintf(w, `{"status": "success", "message": "VALIDATION successful! Your request passed all schema validations."}`)
	} else {
		fmt.Fprintf(w, "VALIDATION successful! Your request passed all schema validations.\n")
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

	log.Printf("Using rules directory: %s", rulesDir)

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

	// Create mux
	mux := http.NewServeMux()
	mux.Handle("/", txhttp.WrapHandler(waf, http.HandlerFunc(successHandler))) // Use the custom handler

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
