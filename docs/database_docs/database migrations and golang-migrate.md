# What Are Database Migrations?

**Database migrations** are a way to incrementally and systematically manage changes to your database schema over time. Think of them as "version control for your database structure" - they allow you to evolve your database schema alongside your application code in a controlled, reversible manner.[1][2][3][4]

## Key Concepts

**Schema Evolution**: Migrations track granular changes like creating tables, adding columns, modifying indexes, or updating constraints.[3][4][1]

**Version Control**: Each migration has a version number and timestamp, creating a chronological history of database changes.[2][5][6]

**Reversibility**: Every migration should have both "up" (apply changes) and "down" (rollback changes) scripts.[7][8][6]

**Environment Consistency**: Ensures your database schema is identical across development, testing, and production environments.[9][1]

## What is golang-migrate?

**golang-migrate** is the most popular database migration tool for Go applications. It provides both a CLI tool for manual operations and a Go library that can be embedded in your applications.[8][6][10][7]

### Features

- **Multi-database support**: PostgreSQL, MySQL, SQLite, SQL Server, and more
- **CLI and library**: Use from command line or integrate into your Go code
- **Version tracking**: Automatically tracks which migrations have been applied
- **Rollback support**: Can revert changes using down migrations
- **Dirty state handling**: Can recover from failed migrations[6][7][8]

## Installation and Setup

### Install CLI Tool

```bash
# Using Go
go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Or using package managers
brew install golang-migrate  # macOS
scoop install migrate        # Windows
```

### Project Structure

```
your-project/
├── migrations/
│   ├── 000001_create_users_table.up.sql
│   ├── 000001_create_users_table.down.sql
│   ├── 000002_add_email_index.up.sql
│   └── 000002_add_email_index.down.sql
├── main.go
└── go.mod
```

## Creating Your First Migration

### Generate Migration Files

```bash
# Create a new migration
migrate create -ext sql -dir migrations -seq create_users_table
```

This creates two files:
- `000001_create_users_table.up.sql` - applies the migration
- `000001_create_users_table.down.sql` - reverts the migration[7][8]

### Up Migration (000001_create_users_table.up.sql)

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

### Down Migration (000001_create_users_table.down.sql)

```sql
DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
```

## Using golang-migrate with Your Mangahub Project

### Integration into Your Go Application

```go
package database

import (
    "database/sql"
    "fmt"
    "log/slog"
    "mangahub/internal/config"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/sqlite3"
    "github.com/golang-migrate/migrate/v4/source/file"
    _ "modernc.org/sqlite"
)

func ConnectDB(cfg *config.Config, logger *slog.Logger) (*sql.DB, error) {
    db, err := sql.Open("sqlite", cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    if err := db.Ping(); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Run migrations
    if err := runMigrations(db, logger); err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to run migrations: %w", err)
    }

    logger.Info("Connected to the database successfully")
    return db, nil
}

func runMigrations(db *sql.DB, logger *slog.Logger) error {
    driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
    if err != nil {
        return fmt.Errorf("failed to create migration driver: %w", err)
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations",
        "sqlite3", 
        driver,
    )
    if err != nil {
        return fmt.Errorf("failed to create migrate instance: %w", err)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return fmt.Errorf("failed to apply migrations: %w", err)
    }

    logger.Info("Database migrations applied successfully")
    return nil
}
```

## Common Migration Commands

### CLI Commands

```bash
# Apply all pending migrations
migrate -path migrations -database "sqlite3://./mangahub.db" up

# Rollback one migration
migrate -path migrations -database "sqlite3://./mangahub.db" down 1

# Check current version
migrate -path migrations -database "sqlite3://./mangahub.db" version

# Force set version (if database is in dirty state)
migrate -path migrations -database "sqlite3://./mangahub.db" force 1
```

## Example Migration Files for Mangahub

### Create Manga Table (000001_create_manga_table.up.sql)

```sql
CREATE TABLE manga (
    id TEXT PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    author VARCHAR(255),
    status VARCHAR(50) DEFAULT 'ongoing',
    cover_image_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_manga_title ON manga(title);
CREATE INDEX idx_manga_author ON manga(author);
CREATE INDEX idx_manga_status ON manga(status);
```

### Down Migration (000001_create_manga_table.down.sql)

```sql
DROP INDEX IF EXISTS idx_manga_status;
DROP INDEX IF EXISTS idx_manga_author;
DROP INDEX IF EXISTS idx_manga_title;
DROP TABLE IF EXISTS manga;
```

## Best Practices

**Always write down migrations**: Every up migration should have a corresponding down migration.[8][6][7]

**Test migrations**: Test both up and down migrations in a development environment before production.[9][7]

**Incremental changes**: Keep migrations small and focused on specific changes.[2][3]

**Backup before migrations**: Always backup production databases before applying migrations.[11][1]

**Version control**: Store migration files in your version control system alongside your code.[5][12][2]

Database migrations are essential for maintaining database schema consistency across environments and enabling safe, reversible database changes as your application evolves. The golang-migrate tool provides a robust, production-ready solution for managing these migrations in Go projects.[1][3][6][7][8][9]

[1](https://www.acronis.com/en/blog/posts/what-is-database-migration/)
[2](https://docs.evolveum.com/midpoint/reference/master/repository/generic/database-schema-versioning/)
[3](https://www.prisma.io/dataguide/types/relational/what-are-database-migrations)
[4](https://www.cloudbees.com/blog/database-migration)
[5](https://stackoverflow.com/questions/175451/how-do-you-version-your-database-schema)
[6](https://betterstack.com/community/guides/scaling-go/golang-migrate/)
[7](https://neon.com/guides/golang-db-migrations-postgres)
[8](https://dev.to/jad_core/database-schema-migration-cheatsheet-with-golang-migratemigrate-35b3)
[9](https://viblo.asia/p/db-migration-for-golang-services-why-it-matters-W13VMWg5JY7)
[10](https://github.com/golang-migrate/migrate)
[11](https://cloud.google.com/architecture/database-migration-concepts-principles-part-1)
[12](https://betterprogramming.pub/keeping-track-of-database-schema-changes-f3a227e29f5f)
[13](https://www.reddit.com/r/node/comments/90fo0t/whats_datadatabase_migration/)
[14](https://atlasgo.io/guides/migration-tools/golang-migrate)
[15](https://www.reddit.com/r/golang/comments/14voypr/database_migration_tool/)
[16](https://viblo.asia/p/managing-database-migration-in-go-aWj5336b56m)
[17](https://blog.jetbrains.com/idea/2025/02/database-migrations-in-the-real-world/)
[18](https://github.com/golang-migrate/migrate/issues/129)
[19](https://www.youtube.com/watch?v=mMsZPZKNc4g)
[20](https://dev.to/ouma_ouma/mastering-database-migrations-in-go-with-golang-migrate-and-sqlite-3jhb)