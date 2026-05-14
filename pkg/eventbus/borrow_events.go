package eventbus

import "time"

type BookBorrowedEvent struct {
	BorrowID string    `json:"borrow_id"`
	UserID   string    `json:"user_id"`
	BookID   string    `json:"book_id"`
	ExpID    string    `json:"exp_id"`
	DateFrom time.Time `json:"date_from"`
	DateTo   time.Time `json:"date_to"`
}

type BookReturnedEvent struct {
	BorrowID   string    `json:"borrow_id"`
	UserID     string    `json:"user_id"`
	BookID     string    `json:"book_id"`
	ExpID      string    `json:"exp_id"`
	ReturnedAt time.Time `json:"returned_at"`
}

type BookReservedEvent struct {
	BorrowID string    `json:"borrow_id"`
	UserID   string    `json:"user_id"`
	BookID   string    `json:"book_id"`
	ExpID    string    `json:"exp_id"`
	At       time.Time `json:"at"`
}

type ReservationCancelledEvent struct {
	BorrowID string `json:"borrow_id"`
	UserID   string `json:"user_id"`
	ExpID    string `json:"exp_id"`
}