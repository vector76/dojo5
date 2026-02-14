CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    phone TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'instructor', 'user')),
    password_hash TEXT NOT NULL,
    membership_type TEXT CHECK (membership_type IN ('monthly', 'annual', 'drop-in')),
    membership_status TEXT CHECK (membership_status IN ('active', 'inactive', 'suspended')),
    emergency_contact TEXT,
    join_date DATE,
    expected_balance REAL NOT NULL DEFAULT 0,
    deleted_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
