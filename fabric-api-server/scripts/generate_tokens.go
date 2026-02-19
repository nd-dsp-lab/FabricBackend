package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"go-huma-api-server/src/auth"
)

// tokenData represents the JSON structure for token storage
type tokenData struct {
	Tokens map[string]string `json:"tokens"`
}

func main() {
	var (
		count      = flag.Int("n", 5, "Number of tokens to generate")
		outputFile = flag.String("o", "tokens.json", "Output file path for token mappings")
		tokenLen   = flag.Int("token-len", 32, "Length of generated tokens")
		secretLen  = flag.Int("secret-len", 64, "Length of generated secret keys")
	)
	flag.Parse()

	if *count <= 0 {
		fmt.Fprintf(os.Stderr, "Error: count must be positive\n")
		os.Exit(1)
	}

	// Load existing tokens if file exists
	tokens := make(map[string]string)
	if data, err := os.ReadFile(*outputFile); err == nil {
		var existing tokenData
		if err := json.Unmarshal(data, &existing); err == nil {
			tokens = existing.Tokens
			fmt.Printf("Loaded %d existing tokens from %s\n", len(tokens), *outputFile)
		}
	}

	// Generate new tokens
	generated := 0
	for i := 0; i < *count; i++ {
		token := auth.GenerateToken(*tokenLen)
		secret := auth.GenerateSecret(*secretLen)

		// Ensure uniqueness
		if _, exists := tokens[token]; exists {
			// Skip duplicates and try again
			i--
			continue
		}

		tokens[token] = secret
		generated++
	}

	// Save to file
	td := tokenData{Tokens: tokens}
	data, err := json.MarshalIndent(td, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling tokens: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*outputFile, data, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing token file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated %d new token(s)\n", generated)
	fmt.Printf("Total tokens in %s: %d\n", *outputFile, len(tokens))
	fmt.Println("\nGenerated tokens:")

	// Print the newly generated tokens
	i := 0
	for token, secret := range tokens {
		if i >= len(tokens)-generated {
			fmt.Printf("  Token:  %s\n", token)
			fmt.Printf("  Secret: %s\n\n", secret)
		}
		i++
	}
}
