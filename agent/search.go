package agent

import (
	"encoding/json"
	"net/http"

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

func (s *SearchService) Search(w http.ResponseWriter, r *http.Request) {
	// Parse the search query from the request
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "Missing search query", http.StatusBadRequest)
		return
	}

	// Perform the search using the collection
	results, err := s.collection.Query(r.Context(), query, 10, nil, nil)
	if err != nil {
		http.Error(w, "Error performing search: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the search results as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
