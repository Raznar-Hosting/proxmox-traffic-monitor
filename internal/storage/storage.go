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
	ID     string `json:"id"`
	VMID   int64  `json:"vmid"`   // new field
	NodeID string `json:"nodeid"` // new field
	In     uint64 `json:"in"`
	Out    uint64 `json:"out"`

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
		vmid TEXT NOT NULL,
		nodeid TEXT NOT NULL,
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
	INSERT INTO traffic (id, vmid, nodeid, net_in, net_out, timestamp)
	VALUES (?, ?, ?, ?, ?, ?);
	`

	ts := time.Now().Unix()
	if record.Timestamp != nil {
		ts = record.Timestamp.Unix()
	}

	_, err := s.db.Exec(
		query,
		record.ID,
		record.VMID,
		record.NodeID,
		record.In,
		record.Out,
		ts,
	)
	if err != nil {
		return fmt.Errorf("failed to update traffic: %w", err)
	}
	return nil
}

func (s *Storage) GetTraffic(id string) ([]TrafficRecord, error) {
	query := `
		SELECT id, vmid, nodeid,
		       SUM(net_in)  AS net_in,
		       SUM(net_out) AS net_out
		FROM traffic
	`
	args := []any{}

	if id != "" {
		query += " WHERE id = ?"
		args = append(args, id)
	}

	query += " GROUP BY id, vmid, nodeid ORDER BY id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]TrafficRecord, 0)
	for rows.Next() {
		var r TrafficRecord
		if err := rows.Scan(
			&r.ID,
			&r.VMID,
			&r.NodeID,
			&r.In,
			&r.Out,
		); err != nil {
			return nil, err
		}
		// Timestamp intentionally nil: this is an aggregate
		records = append(records, r)
	}

	return records, rows.Err()
}



func (s *Storage) GetDailyTraffic(id string, day time.Time) ([]TrafficRecord, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.AddDate(0, 0, 1) // next day

	query := `
	SELECT id, vmid, nodeid, SUM(net_in) AS net_in, SUM(net_out) AS net_out,
	       strftime('%s', timestamp, 'unixepoch', 'start of day') AS day_ts
	FROM traffic
	WHERE timestamp >= ? AND timestamp < ?
	`
	args := []any{start.Unix(), end.Unix()}

	if id != "" {
		query += " AND id = ?"
		args = append(args, id)
	}

	query += " GROUP BY id, vmid, nodeid, day_ts ORDER BY day_ts, id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]TrafficRecord, 0)
	for rows.Next() {
		var r TrafficRecord
		var dayTS int64
		if err := rows.Scan(&r.ID, &r.VMID, &r.NodeID, &r.In, &r.Out, &dayTS); err != nil {
			return nil, err
		}
		t := time.Unix(dayTS, 0)
		r.Timestamp = &t
		records = append(records, r)
	}

	return records, rows.Err()
}

func (s *Storage) GetMonthlyTraffic(id string, monthYear string) ([]TrafficRecord, error) {
	// monthYear format: YYYY-MM
	start, err := time.Parse("2006-01", monthYear)
	if err != nil {
		return nil, fmt.Errorf("invalid month format: %w", err)
	}
	end := start.AddDate(0, 1, 0) // next month

	query := `
	SELECT id, vmid, nodeid,
	       SUM(net_in) AS net_in,
	       SUM(net_out) AS net_out,
	       strftime('%s', timestamp, 'unixepoch', 'start of month') AS month_ts
	FROM traffic
	WHERE timestamp >= ? AND timestamp < ?
	`
	args := []any{start.Unix(), end.Unix()}

	if id != "" {
		query += " AND id = ?"
		args = append(args, id)
	}

	query += " GROUP BY id, vmid, nodeid, month_ts ORDER BY id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]TrafficRecord, 0)
	for rows.Next() {
		var r TrafficRecord
		var monthTS int64
		if err := rows.Scan(&r.ID, &r.VMID, &r.NodeID, &r.In, &r.Out, &monthTS); err != nil {
			return nil, err
		}
		t := time.Unix(monthTS, 0)
		r.Timestamp = &t
		records = append(records, r)
	}

	return records, rows.Err()
}


func (s *Storage) GetTrafficByRange(id string, start, end *time.Time) ([]TrafficRecord, error) {
	// Default end = now
	if end == nil {
		now := time.Now()
		end = &now
	}

	// Default start = very old (Unix epoch)
	if start == nil {
		t := time.Unix(0, 0)
		start = &t
	}

	query := `
		SELECT id, vmid, nodeid, net_in, net_out, timestamp
		FROM traffic
		WHERE timestamp BETWEEN ? AND ?
	`
	args := []any{start.Unix(), end.Unix()}

	if id != "" {
		query += " AND id = ?"
		args = append(args, id)
	}

	query += " ORDER BY timestamp, id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]TrafficRecord, 0)
	for rows.Next() {
		var r TrafficRecord
		var ts int64

		if err := rows.Scan(
			&r.ID,
			&r.VMID,
			&r.NodeID,
			&r.In,
			&r.Out,
			&ts,
		); err != nil {
			return nil, err
		}

		t := time.Unix(ts, 0)
		r.Timestamp = &t

		records = append(records, r)
	}

	return records, rows.Err()
}

func (s *Storage) GetTrafficTotalByRange(
	id string,
	start, end *time.Time,
) ([]TrafficRecord, error) {

	// Default end = now
	if end == nil {
		now := time.Now()
		end = &now
	}

	// Default start = unix epoch
	if start == nil {
		t := time.Unix(0, 0)
		start = &t
	}

	query := `
	SELECT id, vmid, nodeid,
	       SUM(net_in)  AS total_in,
	       SUM(net_out) AS total_out
	FROM traffic
	WHERE timestamp >= ? AND timestamp <= ?`
	args := []interface{}{
		start.Unix(),
		end.Unix(),
	}

	if id != "" {
		query += " AND id = ?"
		args = append(args, id)
	}

	query += " GROUP BY id, vmid, nodeid ORDER BY id"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]TrafficRecord, 0)
	for rows.Next() {
		var r TrafficRecord
		if err := rows.Scan(
			&r.ID,
			&r.VMID,
			&r.NodeID,
			&r.In,
			&r.Out,
		); err != nil {
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
	// cutoff = now - N days
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour).Unix()

	query := `
	DELETE FROM traffic
	WHERE timestamp < ?;
	`

	res, err := s.db.Exec(query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old records: %w", err)
	}

	return res.RowsAffected()
}

// ExistsTraffic returns true if there is at least one traffic record for the given VM ID.
// If id is empty, it checks if there is any traffic record at all.
func (s *Storage) ExistsTraffic(id string) (bool, error) {
	var query string
	var args []interface{}

	if id != "" {
		query = "SELECT 1 FROM traffic WHERE id = ? LIMIT 1"
		args = append(args, id)
	} else {
		query = "SELECT 1 FROM traffic LIMIT 1"
	}

	row := s.db.QueryRow(query, args...)
	var dummy int
	err := row.Scan(&dummy)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
