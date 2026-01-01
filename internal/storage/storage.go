package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/sqlite"
)

type TrafficRecord struct {
	ID        string     `json:"id"`        
	Date      string     `json:"date"`      
	In        uint64     `json:"in"`        
	Out       uint64     `json:"out"`       
	Timestamp *time.Time `json:"timestamp,omitempty"` // omit if nil
}

type Storage struct {
	db *sql.DB
}

func New(dbPath string) (*Storage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Storage) init() error {
	query := `
	CREATE TABLE IF NOT EXISTS traffic (
		id TEXT NOT NULL,
		date TEXT NOT NULL,
		net_in INTEGER NOT NULL,
		net_out INTEGER NOT NULL,
		timestamp INTEGER NOT NULL,
		PRIMARY KEY (id, timestamp)
	);`
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func (s *Storage) UpdateTraffic(record TrafficRecord) error {
	query := `
	INSERT INTO traffic (id, date, net_in, net_out, timestamp)
	VALUES (?, ?, ?, ?, ?);
	`
	_, err := s.db.Exec(query, record.ID, record.Date, record.In, record.Out, record.Timestamp.Unix())
	if err != nil {
		return fmt.Errorf("failed to update traffic: %w", err)
	}
	return nil
}
func (s *Storage) GetTraffic(id string) ([]TrafficRecord, error) {
	var query string
	var args []interface{}

	if id != "" {
		// Sum all traffic for a specific VM
		query = `
			SELECT id, SUM(net_in) as net_in, SUM(net_out) as net_out
			FROM traffic
			WHERE id = ?
			GROUP BY id
		`
		args = append(args, id)
	} else {
		// Sum traffic for all VMs
		query = `
			SELECT id, SUM(net_in) as net_in, SUM(net_out) as net_out
			FROM traffic
			GROUP BY id
			ORDER BY id
		`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []TrafficRecord
	for rows.Next() {
		var r TrafficRecord
		if id != "" {
			// For all-time totals, date field can be empty or "ALL"
			r.Date = "ALL"
		} else {
			r.Date = "ALL"
		}
		if err := rows.Scan(&r.ID, &r.In, &r.Out); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func (s *Storage) GetDailyTraffic(id string, date string) ([]TrafficRecord, error) {
	var query string
	var args []interface{}

	if id != "" {
		query = "SELECT id, date, net_in, net_out FROM traffic WHERE id = ? AND date = ?"
		args = append(args, id, date)
	} else {
		query = "SELECT id, date, net_in, net_out FROM traffic WHERE date = ?"
		args = append(args, date)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []TrafficRecord
	for rows.Next() {
		var r TrafficRecord
		if err := rows.Scan(&r.ID, &r.Date, &r.In, &r.Out); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func (s *Storage) GetMonthlyTraffic(id string, monthYear string) ([]TrafficRecord, error) {
	// monthYear should be -MM-YY
	var query string
	var args []interface{}

	if id != "" {
		query = `
		SELECT id, SUBSTR(date, 4) as month, SUM(net_in), SUM(net_out)
		FROM traffic
		WHERE id = ? AND date LIKE ?
		GROUP BY id, month`
		args = append(args, id, "%"+monthYear)
	} else {
		query = `
		SELECT id, SUBSTR(date, 4) as month, SUM(net_in), SUM(net_out)
		FROM traffic
		WHERE date LIKE ?
		GROUP BY id, month`
		args = append(args, "%"+monthYear)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []TrafficRecord
	for rows.Next() {
		var r TrafficRecord
		if err := rows.Scan(&r.ID, &r.Date, &r.In, &r.Out); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func (s *Storage) ClearTraffic(id string) error {
	var query string
	var args []interface{}

	if id != "" {
		query = "DELETE FROM traffic WHERE id = ?"
		args = append(args, id)
	} else {
		query = "DELETE FROM traffic"
	}

	_, err := s.db.Exec(query, args...)
	return err
}

func (s *Storage) CleanupOldRecords(days int) (int64, error) {
	query := `
	DELETE FROM traffic
	WHERE
		-- Reorder DD-MM-YY to YY-MM-DD for correct date comparison
		('20' || SUBSTR(date, 7, 2) || '-' || SUBSTR(date, 4, 2) || '-' || SUBSTR(date, 1, 2))
		< date('now', '-' || ? || ' days')
	`
	res, err := s.db.Exec(query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old records: %w", err)
	}

	return res.RowsAffected()
}

func (s *Storage) Close() error {
	return s.db.Close()
}
