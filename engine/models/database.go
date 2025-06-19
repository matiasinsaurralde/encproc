package models

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type EncProc struct {
	ID     string
	Pk     []byte
	Params string
}

type EncProcModel struct {
	DB *sql.DB
}

type EncProcModelAPI interface {
	InitializeTables()
	InsertAggregationParams(id string, pk []byte, params string) error
	UpdateAggregationParams(id string, pk []byte, params string) error
	DeleteAggregationParams(id string) error
	GetAggregationParamsByID(id string) (string, []byte, string, error)
	IDexists(id string) (bool, error)
	InsertAggregation(id string, ctAggr []byte, sampleSize int) error
	GetAggregationsByID(id string) ([]struct {
		ID         string
		CtAggr     []byte
		SampleSize int
		CreatedAt  string
	}, error)
	DeleteAggregation(id string) error
}

// initializeTables creates the required tables if they do not exist.
func (m *EncProcModel) InitializeTables() {
	// Table: AggregationParams
	aggregationParamsQuery := `
	CREATE TABLE IF NOT EXISTS AggregationParams (
		id VARCHAR(255),
		pk MEDIUMBLOB,
		params VARCHAR(255),
		PRIMARY KEY (id)
	);`
	_, err := m.DB.Exec(aggregationParamsQuery)
	if err != nil {
		log.Fatalf("Failed to create AggregationParams table: %v\n", err)
	}

	// Table: Aggregation
	aggregationQuery := `
	CREATE TABLE IF NOT EXISTS Aggregation (
		id VARCHAR(255) NOT NULL,
		ct_aggr MEDIUMBLOB NOT NULL,
		sample_size INT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`
	_, err = m.DB.Exec(aggregationQuery)
	if err != nil {
		log.Fatalf("Failed to create Aggregation table: %v\n", err)
	}
	log.Println("Database tables initialized successfully")
}

// InsertAggregationParams inserts a new entry into the AggregationParams table.
func (m *EncProcModel) InsertAggregationParams(id string, pk []byte, params string) error {
	query := "INSERT INTO AggregationParams (id, pk, params) VALUES (?, ?, ?)"
	_, err := m.DB.Exec(query, id, pk, params)
	return err
}

// UpdateAggregationParams updates an existing entry in the AggregationParams table.
func (m *EncProcModel) UpdateAggregationParams(id string, pk []byte, params string) error {
	query := "UPDATE AggregationParams SET pk = ?, params = ? WHERE id = ?"
	result, err := m.DB.Exec(query, pk, params, id)
	if err != nil {
		return err
	}

	// Check if any rows were affected
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	/*
		if rowsAffected == 0 {
			return fmt.Errorf("No entry found with ID: %s", id)
		}
	*/
	return nil
}

// DeleteAggregationParams deletes an entry from the AggregationParams table by ID.
func (m *EncProcModel) DeleteAggregationParams(id string) error {
	query := "DELETE FROM AggregationParams WHERE id = ?"
	_, err := m.DB.Exec(query, id)
	return err
}

// GetAggregationParamsByID retrieves an entry from the AggregationParams table by its ID.
func (m *EncProcModel) GetAggregationParamsByID(id string) (string, []byte, string, error) {
	query := "SELECT id, pk, params FROM AggregationParams WHERE id = ?"

	// Define variables to hold the retrieved data
	var retrievedID string
	var pk []byte
	var params string

	// Execute the query
	err := m.DB.QueryRow(query, id).Scan(&retrievedID, &pk, &params)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil, "", fmt.Errorf("no entry found for ID: %s", id)
		}
		return "", nil, "", err
	}

	return retrievedID, pk, params, nil
}

// IDexists checks whether the given id exists in the AggregationParams table.
func (m *EncProcModel) IDexists(id string) (bool, error) {
	query := "SELECT 1 FROM AggregationParams WHERE id = ? LIMIT 1"

	var exists int
	err := m.DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// InsertAggregation inserts a new entry into the Aggregation table.
func (m *EncProcModel) InsertAggregation(id string, ctAggr []byte, sampleSize int) error {
	query := "INSERT INTO Aggregation (id, ct_aggr, sample_size) VALUES (?, ?, ?)"
	_, err := m.DB.Exec(query, id, ctAggr, sampleSize)
	return err
}

// GetAggregationsByID retrieves all entries from the Aggregation table with the given ID.
func (m *EncProcModel) GetAggregationsByID(id string) ([]struct {
	ID         string
	CtAggr     []byte
	SampleSize int
	CreatedAt  string
}, error) {
	query := "SELECT id, ct_aggr, sample_size, created_at FROM Aggregation WHERE id = ?"

	// Slice to hold results
	var aggregations []struct {
		ID         string
		CtAggr     []byte
		SampleSize int
		CreatedAt  string
	}

	// Execute the query
	rows, err := m.DB.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the result set
	for rows.Next() {
		var aggregation struct {
			ID         string
			CtAggr     []byte
			SampleSize int
			CreatedAt  string
		}
		err := rows.Scan(&aggregation.ID, &aggregation.CtAggr, &aggregation.SampleSize, &aggregation.CreatedAt)
		if err != nil {
			return nil, err
		}
		aggregations = append(aggregations, aggregation)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return aggregations, nil
}

// DeleteAggregation deletes an entry from the Aggregation table by ID.
func (m *EncProcModel) DeleteAggregation(id string) error {
	query := "DELETE FROM Aggregation WHERE id = ?"
	_, err := m.DB.Exec(query, id)
	return err
}
