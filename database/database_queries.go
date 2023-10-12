package database

const UserTableQuery = `CREATE TABLE IF NOT EXISTS users(
	id TEXT PRIMARY KEY,
	email TEXT NOT NULL,
	password TEXT
);`

const CreateTableIfNotExistsQuery = `CREATE TABLE IF NOT EXISTS todo_items (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,
	due_date TIMESTAMP,
	priority TEXT,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	is_completed BOOLEAN DEFAULT FALSE,
	is_deleted BOOLEAN DEFAULT FALSE,
	user_id TEXT REFERENCES users(id)
);`
