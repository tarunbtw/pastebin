package main

import "time"

type Paste struct {
	ID           string     `json:"id"`
	Content      string     `json:"content"`
	Language     string     `json:"language"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	ViewLimit    *int       `json:"view_limit,omitempty"`
	ViewCount    int        `json:"view_count"`
	CreatedAt    time.Time  `json:"created_at"`
	PasswordHash string     `json:"-"`
	Burn         bool       `json:"burn,omitempty"`
}

type CreatePasteRequest struct {
	Content   string `json:"content"`
	Language  string `json:"language"`
	ExpiresIn string `json:"expires_in"`
	ViewLimit *int   `json:"view_limit"`
	Password  string `json:"password"`
	Burn      bool   `json:"burn"`
}

type CreatePasteResponse struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}