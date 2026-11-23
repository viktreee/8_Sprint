package main

import (
	"database/sql"
	"fmt"
)

// ParcelStore stores and retrieves parcels in SQLite.
type ParcelStore struct {
	db *sql.DB
}

// NewParcelStore returns a store backed by db.
func NewParcelStore(db *sql.DB) ParcelStore {
	return ParcelStore{db: db}
}

// Add saves p and returns its assigned number.
func (s ParcelStore) Add(p Parcel) (int, error) {
	result, err := s.db.Exec("INSERT INTO parcel (client, status, address, created_at) VALUES (?, ?, ?, ?)",
		p.Client, p.Status, p.Address, p.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("insert parcel: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get inserted parcel number: %w", err)
	}
	return int(id), nil
}

// Get returns the parcel by number or an error wrapping sql.ErrNoRows if absent.
func (s ParcelStore) Get(number int) (Parcel, error) {
	var p Parcel
	err := s.db.QueryRow("SELECT number, client, status, address, created_at FROM parcel WHERE number = ?", number).
		Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt)
	if err != nil {
		return Parcel{}, fmt.Errorf("get parcel %d: %w", number, err)
	}
	return p, nil
}

// GetByClient returns the client's parcels in tracking number order.
func (s ParcelStore) GetByClient(client int) ([]Parcel, error) {
	rows, err := s.db.Query("SELECT number, client, status, address, created_at FROM parcel WHERE client = ? ORDER BY number", client)
	if err != nil {
		return nil, fmt.Errorf("list parcels for client %d: %w", client, err)
	}
	defer rows.Close()

	var parcels []Parcel
	for rows.Next() {
		var p Parcel
		if err := rows.Scan(&p.Number, &p.Client, &p.Status, &p.Address, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("read parcel: %w", err)
		}
		parcels = append(parcels, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate parcels: %w", err)
	}
	return parcels, nil
}

// SetStatus sets the status of the parcel identified by number.
func (s ParcelStore) SetStatus(number int, status string) error {
	_, err := s.db.Exec("UPDATE parcel SET status = ? WHERE number = ?", status, number)
	if err != nil {
		return fmt.Errorf("update parcel status %d: %w", number, err)
	}
	return nil
}

// SetAddress changes the address only while the parcel is registered.
func (s ParcelStore) SetAddress(number int, address string) error {
	_, err := s.db.Exec("UPDATE parcel SET address = ? WHERE number = ? AND status = ?", address, number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("update parcel address %d: %w", number, err)
	}
	return nil
}

// Delete removes only registered parcels.
func (s ParcelStore) Delete(number int) error {
	_, err := s.db.Exec("DELETE FROM parcel WHERE number = ? AND status = ?", number, ParcelStatusRegistered)
	if err != nil {
		return fmt.Errorf("delete parcel %d: %w", number, err)
	}
	return nil
}
