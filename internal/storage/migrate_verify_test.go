package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

func TestDropLegacyDateColumn(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "traffic.db")

	s, err := New(dbPath)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	// Simulate the legacy schema by adding the unused `date` column.
	if _, err := s.db.Exec("ALTER TABLE traffic ADD COLUMN date TEXT NOT NULL DEFAULT ''"); err != nil {
		t.Fatalf("add legacy date column: %v", err)
	}

	if err := s.dropLegacyDateColumn(); err != nil {
		t.Fatalf("dropLegacyDateColumn: %v", err)
	}

	// Inserts must now succeed without the date column.
	now := time.Now()
	if err := s.UpdateTraffic(TrafficRecord{
		ID: "node-100-vm", VMID: 100, NodeID: "node",
		In: 123, Out: 456, Timestamp: &now,
	}); err != nil {
		t.Fatalf("UpdateTraffic after migration: %v", err)
	}

	// Verify the date column is gone.
	rows, err := s.db.Query("PRAGMA table_info(traffic)")
	if err != nil {
		t.Fatalf("pragma: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		if name == "date" {
			t.Fatal("date column still exists")
		}
	}
}
