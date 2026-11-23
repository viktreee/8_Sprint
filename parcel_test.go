package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// getTestParcel returns a registered parcel for testing.
func getTestParcel() Parcel {
	return Parcel{
		Client:    1000,
		Status:    ParcelStatusRegistered,
		Address:   "test",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// openTestDB opens a temporary copy of tracker.db for the test.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	data, err := os.ReadFile("tracker.db")
	require.NoError(t, err)
	path := filepath.Join(t.TempDir(), "tracker.db")
	require.NoError(t, os.WriteFile(path, data, 0600))
	db, err := sql.Open("sqlite", path)
	require.NoError(t, err)
	return db
}

// TestAddGetDelete checks that a parcel can be saved, retrieved, and deleted.
func TestAddGetDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.True(t, id > 0)
	parcel.Number = id

	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, parcel, stored)

	require.NoError(t, store.Delete(id))
	_, err = store.Get(id)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestSetAddress checks that a registered parcel's address can change.
func TestSetAddress(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.True(t, id > 0)

	newAddress := "new test address"
	require.NoError(t, store.SetAddress(id, newAddress))
	stored, err := store.Get(id)
	require.NoError(t, err)
	require.Equal(t, newAddress, stored.Address)
}

// TestSetStatus checks that sent and delivered parcels retain their address and cannot be deleted.
func TestSetStatus(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)
	parcel := getTestParcel()
	id, err := store.Add(parcel)
	require.NoError(t, err)
	require.True(t, id > 0)
	parcel.Number = id

	for _, status := range []string{ParcelStatusSent, ParcelStatusDelivered} {
		require.NoError(t, store.SetStatus(id, status))
		require.NoError(t, store.SetAddress(id, "another address"))
		require.NoError(t, store.Delete(id))
		stored, err := store.Get(id)
		require.NoError(t, err)
		parcel.Status = status
		require.Equal(t, parcel, stored)
	}
}

// TestGetByClient checks that all and only the requested client's parcels are returned.
func TestGetByClient(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	store := NewParcelStore(db)
	client := 1000
	_, err := db.Exec("DELETE FROM parcel WHERE client = ?", client)
	require.NoError(t, err)
	parcelMap := map[int]Parcel{}
	for i := 0; i < 3; i++ {
		parcel := getTestParcel()
		parcel.Client = client
		id, err := store.Add(parcel)
		require.NoError(t, err)
		require.True(t, id > 0)
		parcel.Number = id
		parcelMap[id] = parcel
	}
	other := getTestParcel()
	other.Client = client + 1
	_, err = store.Add(other)
	require.NoError(t, err)

	storedParcels, err := store.GetByClient(client)
	require.NoError(t, err)
	require.Len(t, storedParcels, len(parcelMap))
	for _, parcel := range storedParcels {
		expected, ok := parcelMap[parcel.Number]
		require.True(t, ok)
		require.Equal(t, expected, parcel)
		delete(parcelMap, parcel.Number)
	}
}
