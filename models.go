package main

import "time"

type Paste struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	Language  string    `json:"language"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	ViewLimit *int      `json:"view_limit,omitempty"`
	ViewCount int       `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatePasteRequest struct {
	Content   string `json:"content"`
	Language  string `json:"language"`
	ExpiresIn string `json:"expires_in"` // "1h", "24h", "7d", "never"
	ViewLimit *int   `json:"view_limit"`
}

type CreatePasteResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}