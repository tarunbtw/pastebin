## A simple pastebin in Golang


### Folder Structure

```
pastebin/
├── static/
│   ├── index.html       # UI: Create new paste
│   └── paste.html       # UI: View existing paste
├── db.go                # DB connection & SQL queries
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── handlers.go          # HTTP route handlers & logic
├── main.go              # Application entry point & server boot
├── middleware.go        # Logging, recovery, & CORS logic
└── models.go            # Data structures (Paste, API responses)
```