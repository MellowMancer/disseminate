package repositories

import (
	repo_supabase "backend/repositories/supabase"
	"io"
	"net/http"
)

// NewRequest creates a new HTTP request with Supabase authentication headers
// This is a package-level function that can be called directly after importing the package
func NewRequest(supabaseRepository *repo_supabase.SupabaseRepository, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", supabaseRepository.SupabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseRepository.SupabaseKey)

	if method != "GET" {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Prefer", "return=representation")
	}
	return req, nil
}
