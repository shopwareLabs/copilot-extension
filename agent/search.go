package agent

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/philippgille/chromem-go"
)

type SearchService struct {
	collection *chromem.Collection
}

func NewSearchService(collection *chromem.Collection) *SearchService {
	return &SearchService{
		collection: collection,
	}
}

type SearchResult struct {
	ID         string  `json:"id"`
	Similarity float32 `json:"similarity"`
	Content    string  `json:"content"`
}

func (s *SearchService) Search(w http.ResponseWriter, r *http.Request) {
	// Parse the search query from the request
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10"
	}

	// Convert limit to an integer
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		http.Error(w, "Invalid limit value", http.StatusBadRequest)
		return
	}

	// Perform the search using the collection
	results, err := s.collection.Query(r.Context(), query, limitInt, map[string]string{
		"source": "docs",
	}, nil)
	if err != nil {
		http.Error(w, "Error performing search: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the results to a slice of SearchResult
	var searchResults []SearchResult
	for _, result := range results {
		searchResults = append(searchResults, SearchResult{
			ID:         result.ID,
			Similarity: result.Similarity,
			Content:    result.Content,
		})
	}

	// Return the search results as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(searchResults); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
