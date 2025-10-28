package supabase

import (
	"net/http"
)

type SupabaseRepository struct {
	SupabaseKey string
	SupabaseURL string
	HttpClient *http.Client
}

func NewSupabaseRepository(supabaseUrl, supabaseKey string) *SupabaseRepository {
	return &SupabaseRepository{
		SupabaseKey: supabaseKey,
		SupabaseURL: supabaseUrl,
		HttpClient: &http.Client{},
	}
}