CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS borrows (
    borrow_id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL,
    book_id    UUID NOT NULL,
    exp_id     UUID,                    -- заполняется на Day 2 (BorrowSpecificCopy)
    barcode    TEXT NOT NULL,
    date_from  TIMESTAMP NOT NULL DEFAULT NOW(),
    date_to    TIMESTAMP NOT NULL,
    status     TEXT NOT NULL DEFAULT 'ACTIVE',  -- ACTIVE / RETURNED / OVERDUE / RESERVED / CANCELLED
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_borrows_user_id ON borrows(user_id);
CREATE INDEX IF NOT EXISTS idx_borrows_book_id ON borrows(book_id);
CREATE INDEX IF NOT EXISTS idx_borrows_status ON borrows(status);
CREATE INDEX IF NOT EXISTS idx_borrows_date_to ON borrows(date_to);