package main

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func servePastePage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/paste.html")
}

func createPasteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreatePasteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request"}`, http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		http.Error(w, `{"error":"content cannot be empty"}`, http.StatusBadRequest)
		return
	}

	if req.Language == "" {
		req.Language = "plaintext"
	}

	paste := &Paste{
		ID:        xid.New().String(),
		Content:   req.Content,
		Language:  req.Language,
		CreatedAt: time.Now(),
		ViewLimit: req.ViewLimit,
		Burn:      req.Burn,
	}

	// burn after read = view limit of 1
	if req.Burn {
		one := 1
		paste.ViewLimit = &one
	}

	// parse expiry
	switch req.ExpiresIn {
	case "1h":
		t := time.Now().Add(1 * time.Hour)
		paste.ExpiresAt = &t
	case "24h":
		t := time.Now().Add(24 * time.Hour)
		paste.ExpiresAt = &t
	case "7d":
		t := time.Now().Add(7 * 24 * time.Hour)
		paste.ExpiresAt = &t
	case "30d":
		t := time.Now().Add(30 * 24 * time.Hour)
		paste.ExpiresAt = &t
	}

	// hash password if provided
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
			return
		}
		paste.PasswordHash = string(hash)
	}

	if err := insertPaste(paste); err != nil {
		http.Error(w, `{"error":"failed to save paste"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CreatePasteResponse{
		ID:  paste.ID,
		URL: "/p/" + paste.ID,
	})
}

func getPasteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/paste/")
	if id == "" {
		http.Error(w, `{"error":"missing paste id"}`, http.StatusBadRequest)
		return
	}

	password := r.Header.Get("X-Paste-Password")

	paste, err := getPaste(id)
	if err == sql.ErrNoRows {
		http.Error(w, `{"error":"paste not found"}`, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	// check expiry
	if paste.ExpiresAt != nil && time.Now().After(*paste.ExpiresAt) {
		deletePaste(id)
		http.Error(w, `{"error":"paste has expired"}`, http.StatusGone)
		return
	}

	// check view limit
	if paste.ViewLimit != nil && paste.ViewCount >= *paste.ViewLimit {
		deletePaste(id)
		http.Error(w, `{"error":"paste has reached view limit"}`, http.StatusGone)
		return
	}

	// check password
	if paste.PasswordHash != "" {
		if password == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":              "password required",
				"password_protected": true,
			})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(paste.PasswordHash), []byte(password)); err != nil {
			_ = subtle.ConstantTimeCompare([]byte(password), []byte(password))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":              "wrong password",
				"password_protected": true,
			})
			return
		}
	}

	// don't count view if creator is previewing
	isPreview := r.URL.Query().Get("preview") == "1"

	if !isPreview {
		incrementViewCount(id)
		if paste.Burn {
			defer deletePaste(id)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(paste)
}

func getRawHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/raw/")
	if id == "" {
		http.NotFound(w, r)
		return
	}

	paste, err := getPaste(id)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if paste.ExpiresAt != nil && time.Now().After(*paste.ExpiresAt) {
		deletePaste(id)
		http.Error(w, "paste has expired", http.StatusGone)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(paste.Content))
}

func startCleanupWorker() {
	for {
		time.Sleep(1 * time.Hour)
		n, err := deleteExpiredPastes()
		if err != nil {
			return
		}
		if n > 0 {
			println("cleaned up", n, "expired pastes")
		}
	}
}