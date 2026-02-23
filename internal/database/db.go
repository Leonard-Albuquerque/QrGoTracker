	package database

	import (
		"database/sql"
		"io/fs"
		"log"
		"os"
		"path/filepath"
		"runtime"

		_ "modernc.org/sqlite"

		"qr-tracker/internal/config"
	)

	func MustConnect(cfg *config.Config) *sql.DB {
		if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
			log.Fatalf("failed create db dir: %v", err)
		}
		db, err := sql.Open("sqlite", cfg.DBPath)
		if err != nil {
			log.Fatalf("open db: %v", err)
		}
		// pragmas
		if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
			log.Fatalf("pragma fk: %v", err)
		}
		runMigrations(db)
		return db
	}

	func runMigrations(db *sql.DB) {
		dir := findMigrationsDir()
		if dir == "" {
			log.Printf("migrations: directory not found, skipping migrations")
			return
		}
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			data, rerr := os.ReadFile(path)
			if rerr != nil {
				return rerr
			}
			if len(data) == 0 {
				return nil
			}
			if _, e := db.Exec(string(data)); e != nil {
				return e
			}
			return nil
		})
		if err != nil {
			log.Fatalf("migrations failed: %v", err)
		}
	}

	// findMigrationsDir attempts to locate the migrations directory by checking
	// the working directory, parent directories up to 5 levels, and the
	// executable directory. Returns the path to migrations or empty string.
	func findMigrationsDir() string {
		// check cwd and parents
		cwd, err := os.Getwd()
		if err == nil {
			p := cwd
			for i := 0; i < 6; i++ {
				cand := filepath.Join(p, "migrations")
				fi, ferr := os.Stat(cand)
				if ferr == nil && fi.IsDir() {
					return cand
				}
				parent := filepath.Dir(p)
				if parent == p {
					break
				}
				p = parent
			}
		}

		// check executable directory
		if exe, err := os.Executable(); err == nil {
			exedir := filepath.Dir(exe)
			cand := filepath.Join(exedir, "../migrations")
			cand = filepath.Clean(cand)
			if fi, ferr := os.Stat(cand); ferr == nil && fi.IsDir() {
				return cand
			}
			// also check same dir
			cand2 := filepath.Join(exedir, "migrations")
			if fi, ferr := os.Stat(cand2); ferr == nil && fi.IsDir() {
				return cand2
			}
		}

		// fallback: check relative to source (GOROOT maybe) using runtime.Caller
		if _, file, _, ok := runtime.Caller(0); ok {
			srcDir := filepath.Dir(file)
			cand := filepath.Join(srcDir, "../../migrations")
			cand = filepath.Clean(cand)
			if fi, ferr := os.Stat(cand); ferr == nil && fi.IsDir() {
				return cand
			}
		}

		return ""
	}
