package eventbus

import "time"

// BookCreatedEvent — payload subject="book.created"
type BookCreatedEvent struct {
	BookID    string    `json:"book_id"`
	Name      string    `json:"name"`
	Authors   string    `json:"authors"`
	Year      int       `json:"year"`
	CreatedAt time.Time `json:"created_at"`
}

// BookUpdatedEvent — payload subject="book.updated"
type BookUpdatedEvent struct {
	BookID  string `json:"book_id"`
	Name    string `json:"name"`
	Authors string `json:"authors"`
	Year    int    `json:"year"`
}

// BookDeletedEvent — payload subject="book.deleted"
type BookDeletedEvent struct {
	BookID string `json:"book_id"`
}

// CopyStatusChangedEvent — payload subject="book.copy.status_changed"
type CopyStatusChangedEvent struct {
	ExpID     string    `json:"exp_id"`
	BookID    string    `json:"book_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	ChangedAt time.Time `json:"changed_at"`
}
