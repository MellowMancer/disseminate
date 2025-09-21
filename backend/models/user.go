package models

import (
	
)

type User struct {
    Email    string `json:"email"`
    Password string `json:"password"` // Store hashed password or use Supabase Auth
}
