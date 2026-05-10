CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS books (
                                     book_id    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       TEXT NOT NULL,
    authors    TEXT[] NOT NULL DEFAULT '{}',
    year       INT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS book_copies (
                                           exp_id     UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    book_id    UUID NOT NULL REFERENCES books(book_id) ON DELETE CASCADE,
    status     TEXT NOT NULL DEFAULT 'AVAILABLE',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );