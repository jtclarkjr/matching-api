package database

import (
	"database/sql"
	"fmt"
	"log"
)

// Migration represents a database migration
type Migration struct {
	Version string
	Up      string
	Down    string
}

// GetMigrations returns all database migrations
func GetMigrations() []Migration {
	return []Migration{
		{
			Version: "001_create_users_table",
			Up: `
				CREATE TABLE IF NOT EXISTS users (
					id UUID PRIMARY KEY,
					email VARCHAR(255) UNIQUE NOT NULL,
					password VARCHAR(255) NOT NULL,
					first_name VARCHAR(100) NOT NULL,
					last_name VARCHAR(100) NOT NULL,
					age INTEGER NOT NULL CHECK (age >= 18 AND age <= 100),
					bio TEXT,
					gender VARCHAR(20) NOT NULL CHECK (gender IN ('male', 'female', 'non-binary')),
					latitude DECIMAL(10, 8),
					longitude DECIMAL(11, 8),
					city VARCHAR(100),
					state VARCHAR(100),
					country VARCHAR(100),
					is_active BOOLEAN DEFAULT true,
					last_seen TIMESTAMP WITH TIME ZONE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
				CREATE INDEX IF NOT EXISTS idx_users_location ON users(latitude, longitude);
				CREATE INDEX IF NOT EXISTS idx_users_age ON users(age);
				CREATE INDEX IF NOT EXISTS idx_users_gender ON users(gender);
				CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
			`,
			Down: `DROP TABLE IF EXISTS users CASCADE;`,
		},
		{
			Version: "002_create_photos_table",
			Up: `
				CREATE TABLE IF NOT EXISTS photos (
					id UUID PRIMARY KEY,
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					url VARCHAR(500) NOT NULL,
					position INTEGER NOT NULL DEFAULT 1,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(user_id, position)
				);
				CREATE INDEX IF NOT EXISTS idx_photos_user_id ON photos(user_id);
				CREATE INDEX IF NOT EXISTS idx_photos_position ON photos(user_id, position);
			`,
			Down: `DROP TABLE IF EXISTS photos CASCADE;`,
		},
		{
			Version: "003_create_preferences_table",
			Up: `
				CREATE TABLE IF NOT EXISTS user_preferences (
					id UUID PRIMARY KEY,
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					age_min INTEGER NOT NULL DEFAULT 18,
					age_max INTEGER NOT NULL DEFAULT 99,
					max_distance INTEGER NOT NULL DEFAULT 50,
					interested_in TEXT[] NOT NULL DEFAULT ARRAY['female'],
					show_me VARCHAR(20) NOT NULL DEFAULT 'everyone',
					only_verified BOOLEAN DEFAULT false,
					hide_distance BOOLEAN DEFAULT false,
					hide_age BOOLEAN DEFAULT false,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(user_id)
				);
				CREATE INDEX IF NOT EXISTS idx_preferences_user_id ON user_preferences(user_id);
			`,
			Down: `DROP TABLE IF EXISTS user_preferences CASCADE;`,
		},
		{
			Version: "004_create_swipes_table",
			Up: `
				CREATE TABLE IF NOT EXISTS swipes (
					id UUID PRIMARY KEY,
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					target_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					action VARCHAR(10) NOT NULL CHECK (action IN ('left', 'right', 'super')),
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(user_id, target_id)
				);
				CREATE INDEX IF NOT EXISTS idx_swipes_user_id ON swipes(user_id);
				CREATE INDEX IF NOT EXISTS idx_swipes_target_id ON swipes(target_id);
				CREATE INDEX IF NOT EXISTS idx_swipes_action ON swipes(action);
			`,
			Down: `DROP TABLE IF EXISTS swipes CASCADE;`,
		},
		{
			Version: "005_create_matches_table",
			Up: `
				CREATE TABLE IF NOT EXISTS matches (
					id UUID PRIMARY KEY,
					user1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					user2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					is_active BOOLEAN DEFAULT true,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(user1_id, user2_id),
					CHECK (user1_id != user2_id)
				);
				CREATE INDEX IF NOT EXISTS idx_matches_user1 ON matches(user1_id);
				CREATE INDEX IF NOT EXISTS idx_matches_user2 ON matches(user2_id);
				CREATE INDEX IF NOT EXISTS idx_matches_active ON matches(is_active);
			`,
			Down: `DROP TABLE IF EXISTS matches CASCADE;`,
		},
		{
			Version: "006_create_chats_table",
			Up: `
				CREATE TABLE IF NOT EXISTS chats (
					id UUID PRIMARY KEY,
					match_id UUID NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
					last_message_at TIMESTAMP WITH TIME ZONE,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
					UNIQUE(match_id)
				);
				CREATE INDEX IF NOT EXISTS idx_chats_match_id ON chats(match_id);
				CREATE INDEX IF NOT EXISTS idx_chats_last_message ON chats(last_message_at);
			`,
			Down: `DROP TABLE IF EXISTS chats CASCADE;`,
		},
		{
			Version: "007_create_messages_table",
			Up: `
				CREATE TABLE IF NOT EXISTS messages (
					id UUID PRIMARY KEY,
					chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
					sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					content TEXT NOT NULL,
					message_type VARCHAR(20) DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'gif')),
					is_read BOOLEAN DEFAULT false,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
				CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender_id);
				CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
				CREATE INDEX IF NOT EXISTS idx_messages_unread ON messages(is_read) WHERE is_read = false;
			`,
			Down: `DROP TABLE IF EXISTS messages CASCADE;`,
		},
		{
			Version: "008_create_refresh_tokens_table",
			Up: `
				CREATE TABLE IF NOT EXISTS refresh_tokens (
					id UUID PRIMARY KEY,
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					token VARCHAR(500) NOT NULL UNIQUE,
					expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
					is_revoked BOOLEAN DEFAULT false,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
				CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
				CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);
			`,
			Down: `DROP TABLE IF EXISTS refresh_tokens CASCADE;`,
		},
		{
			Version: "009_create_notifications_table",
			Up: `
				CREATE TABLE IF NOT EXISTS notifications (
					id UUID PRIMARY KEY,
					user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
					type VARCHAR(50) NOT NULL,
					title VARCHAR(255) NOT NULL,
					message TEXT NOT NULL,
					data JSONB,
					is_read BOOLEAN DEFAULT false,
					is_sent BOOLEAN DEFAULT false,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
				CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(is_read) WHERE is_read = false;
				CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);
				CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at);
			`,
			Down: `DROP TABLE IF EXISTS notifications CASCADE;`,
		},
		{
			Version: "010_create_analytics_events_table",
			Up: `
				CREATE TABLE IF NOT EXISTS analytics_events (
					id UUID PRIMARY KEY,
					user_id UUID REFERENCES users(id) ON DELETE CASCADE,
					event_type VARCHAR(100) NOT NULL,
					event_data JSONB,
					session_id VARCHAR(255),
					ip_address INET,
					user_agent TEXT,
					created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
				);
				CREATE INDEX IF NOT EXISTS idx_analytics_user_id ON analytics_events(user_id);
				CREATE INDEX IF NOT EXISTS idx_analytics_event_type ON analytics_events(event_type);
				CREATE INDEX IF NOT EXISTS idx_analytics_created_at ON analytics_events(created_at);
				CREATE INDEX IF NOT EXISTS idx_analytics_session ON analytics_events(session_id);
			`,
			Down: `DROP TABLE IF EXISTS analytics_events CASCADE;`,
		},
	}
}

// RunMigrations runs all pending migrations
func RunMigrations() error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(); err != nil {
		return err
	}

	migrations := GetMigrations()
	
	for _, migration := range migrations {
		if exists, err := migrationExists(migration.Version); err != nil {
			return err
		} else if exists {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Running migration: %s", migration.Version)
		
		// Start transaction
		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction for migration %s: %v", migration.Version, err)
		}

		// Execute migration
		if _, err := tx.Exec(migration.Up); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
			return fmt.Errorf("failed to execute migration %s: %v", migration.Version, err)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO migrations (version, applied_at) VALUES ($1, NOW())", migration.Version); err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.Printf("Failed to rollback transaction: %v", rollbackErr)
			}
			return fmt.Errorf("failed to record migration %s: %v", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %v", migration.Version, err)
		}

		log.Printf("Successfully applied migration: %s", migration.Version)
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table
func createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	_, err := DB.Exec(query)
	return err
}

// migrationExists checks if a migration has already been applied
func migrationExists(version string) (bool, error) {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = $1", version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// RollbackMigration rolls back the last migration (for development)
func RollbackMigration() error {
	var lastVersion string
	err := DB.QueryRow("SELECT version FROM migrations ORDER BY applied_at DESC LIMIT 1").Scan(&lastVersion)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No migrations to rollback")
			return nil
		}
		return err
	}

	migrations := GetMigrations()
	var migrationToRollback *Migration
	for _, m := range migrations {
		if m.Version == lastVersion {
			migrationToRollback = &m
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("migration %s not found", lastVersion)
	}

	log.Printf("Rolling back migration: %s", lastVersion)

	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	// Execute rollback
	if _, err := tx.Exec(migrationToRollback.Down); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("failed to rollback migration %s: %v", lastVersion, err)
	}

	// Remove migration record
	if _, err := tx.Exec("DELETE FROM migrations WHERE version = $1", lastVersion); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			log.Printf("Failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("failed to remove migration record %s: %v", lastVersion, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %s: %v", lastVersion, err)
	}

	log.Printf("Successfully rolled back migration: %s", lastVersion)
	return nil
}