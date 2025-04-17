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
	corazahttp "github.com/corazawaf/coraza/v3/http"
	"github.com/corazawaf/coraza/v3/types"
)

func main() {
	// Get test directory path
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Parse flags
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Initialize Coraza WAF
	waf, err := coraza.NewWAF(coraza.NewWAFConfig().
		WithDirectives(strings.ReplaceAll(fmt.Sprintf(`
			# Basic configuration settings
			SecRuleEngine On
			SecRequestBodyAccess On
			SecResponseBodyAccess On
			SecRequestBodyLimit 10485760
			SecRequestBodyInMemoryLimit 10485760

				# Include validation rules
				Include %s/rules/json_validation.conf
		`, dir), "\t\t\t", "")).
		WithErrorCallback(func(rule types.MatchedRule) {
			log.Printf("[CORAZA ERROR] %s", rule.ErrorLog)
		}))
	if err != nil {
		log.Fatalf("Failed to create WAF: %v", err)
	}

	log.Printf("Coraza WAF initialized with rules from %s/rules/", dir)

	// Create HTTP handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "Only POST requests are allowed\n")
			return
		}

		// Respond with success if we reach this point
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Validation successful! Your request passed all schema validations.\n")
		log.Printf("[SUCCESS] Request with content-type %s passed validation", r.Header.Get("Content-Type"))
	})

	// Create Coraza middleware
	interceptor := corazahttp.WrapHandler(waf, handler)

	// Create mux
	mux := http.NewServeMux()
	mux.Handle("/validate", interceptor)

	// Add route for static files
	mux.HandleFunc("/schemas/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(dir, r.URL.Path))
	})

	// Add informational endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
Schema Validation Test Server

Available Endpoints:
- POST /validate (with JSON content)
- GET /schemas/user.json

To test validation:
1. For JSON: curl -X POST -H "Content-Type: application/json" --data @valid_user.json http://localhost:%d/validate
`, *port)
	})

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
