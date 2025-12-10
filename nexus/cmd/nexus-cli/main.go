package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Catalog struct {
	Services []ServiceEntry `json:"services"`
}

type ServiceEntry struct {
	Namespace   string          `json:"namespace"`
	Method      string          `json:"method"`
	Description string          `json:"description"`
	Parameters  []ParamMetadata `json:"parameters"`
}

type ParamMetadata struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func main() {
	catalogPath := flag.String("catalog", "", "Path to catalog.json (default: attempts local then ~/.nexus/catalog.json)")
	searchParam := flag.String("search-param", "", "Search service by parameter name (supports snake, camel, pascal cases)")
	flag.Parse()

	finalPath := *catalogPath
	if finalPath == "" {
		finalPath = resolveDefaultCatalog()
	}

	data, err := os.ReadFile(finalPath)
	if err != nil {
		fmt.Printf("Error reading catalog: %v\n", err)
		os.Exit(1)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		fmt.Printf("Error parsing catalog: %v\n", err)
		os.Exit(1)
	}

	if *searchParam != "" {
		results := searchByParam(catalog, *searchParam)
		if len(results) == 0 {
			fmt.Println("No services found with that parameter.")
		} else {
			fmt.Printf("Found %d services with parameter '%s':\n", len(results), *searchParam)
			for _, res := range results {
				fmt.Printf("- %s.%s (Found matching param: %s)\n", res.Namespace, res.Method, res.MatchedParam)
			}
		}
	} else {
		// List all by default
		fmt.Println("Available Services:")
		for _, s := range catalog.Services {
			fmt.Printf("- %s.%s: %s\n", s.Namespace, s.Method, s.Description)
		}
	}
}

type SearchResult struct {
	Namespace    string
	Method       string
	MatchedParam string
}

func searchByParam(catalog Catalog, query string) []SearchResult {
	var results []SearchResult
	normalizedQuery := normalize(query)

	for _, svc := range catalog.Services {
		for _, param := range svc.Parameters {
			if normalize(param.Name) == normalizedQuery {
				results = append(results, SearchResult{
					Namespace:    svc.Namespace,
					Method:       svc.Method,
					MatchedParam: param.Name,
				})
				break
			}
		}
	}
	return results
}

// normalize removes underscores and converts to lowercase to match snake, camel, and pascal cases
func normalize(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, "_", ""))
}

func resolveDefaultCatalog() string {
	// 1. Check local directory
	if _, err := os.Stat("catalog.json"); err == nil {
		return "catalog.json"
	}
	// 2. Check ~/.nexus/catalog.json
	home, err := os.UserHomeDir()
	if err == nil {
		globalPath := filepath.Join(home, ".nexus", "catalog.json")
		if _, err := os.Stat(globalPath); err == nil {
			return globalPath
		}
	}
	// Fallback (will likely fail but returns valid path string)
	return "catalog.json"
}
