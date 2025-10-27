package repositories

import (
	"net/http"
)

type SupabaseRepository struct {
	supabaseKey string
	supabaseURL string
	httpClient *http.Client
}

func NewSupabaseRepository(supabaseUrl, supabaseKey string) *SupabaseRepository {
	return &SupabaseRepository{
		supabaseKey: supabaseKey,
		supabaseURL: supabaseUrl,
		httpClient: &http.Client{},
	}
}