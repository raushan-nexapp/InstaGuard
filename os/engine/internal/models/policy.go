// Package models defines NexappOS's core data types and their
// database access methods.
package models

import (
	"database/sql"
	"fmt"
	"time"
)

// Policy represents a single firewall rule.
// A Policy is NexappOS's equivalent of a FortiGate firewall policy.
type Policy struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	SrcInterface string    `json:"src_interface"`
	DstInterface string    `json:"dst_interface"`
	SrcAddress   string    `json:"src_address"`
	DstAddress   string    `json:"dst_address"`
	Protocol     string    `json:"protocol"`
	DstPort      string    `json:"dst_port"`
	Action       string    `json:"action"` // accept | drop | reject
	Enabled      bool      `json:"enabled"`
	Priority     int       `json:"priority"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Validate checks that required fields are present and values are sane.
func (p *Policy) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.SrcInterface == "" {
		return fmt.Errorf("src_interface is required")
	}
	if p.DstInterface == "" {
		return fmt.Errorf("dst_interface is required")
	}
	switch p.Action {
	case "accept", "drop", "reject":
		// valid
	default:
		return fmt.Errorf("action must be one of: accept, drop, reject")
	}
	// Default defaults
	if p.SrcAddress == "" {
		p.SrcAddress = "any"
	}
	if p.DstAddress == "" {
		p.DstAddress = "any"
	}
	if p.Protocol == "" {
		p.Protocol = "any"
	}
	if p.DstPort == "" {
		p.DstPort = "any"
	}
	if p.Priority == 0 {
		p.Priority = 100
	}
	return nil
}

// ---- DB operations ----

// ListPolicies returns every policy, ordered by priority then id.
func ListPolicies(db *sql.DB) ([]Policy, error) {
	rows, err := db.Query(`
		SELECT id, name, src_interface, dst_interface, src_address, dst_address,
		       protocol, dst_port, action, enabled, priority, description,
		       created_at, updated_at
		FROM policies
		ORDER BY priority ASC, id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		var enabledInt int
		var desc sql.NullString
		if err := rows.Scan(
			&p.ID, &p.Name, &p.SrcInterface, &p.DstInterface,
			&p.SrcAddress, &p.DstAddress, &p.Protocol, &p.DstPort,
			&p.Action, &enabledInt, &p.Priority, &desc,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		p.Enabled = enabledInt == 1
		if desc.Valid {
			p.Description = desc.String
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

// GetPolicy fetches a single policy by id.
func GetPolicy(db *sql.DB, id int64) (*Policy, error) {
	var p Policy
	var enabledInt int
	var desc sql.NullString
	err := db.QueryRow(`
		SELECT id, name, src_interface, dst_interface, src_address, dst_address,
		       protocol, dst_port, action, enabled, priority, description,
		       created_at, updated_at
		FROM policies WHERE id = ?
	`, id).Scan(
		&p.ID, &p.Name, &p.SrcInterface, &p.DstInterface,
		&p.SrcAddress, &p.DstAddress, &p.Protocol, &p.DstPort,
		&p.Action, &enabledInt, &p.Priority, &desc,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Enabled = enabledInt == 1
	if desc.Valid {
		p.Description = desc.String
	}
	return &p, nil
}

// CreatePolicy inserts a new policy and fills in p.ID on success.
func CreatePolicy(db *sql.DB, p *Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	enabled := 1
	if !p.Enabled {
		enabled = 0
	}
	// On creation, default enabled=true if not explicitly false
	// (Go bool zero value is false — we treat missing as true)
	// For simplicity here, caller should pass Enabled=true for new policies.

	res, err := db.Exec(`
		INSERT INTO policies
			(name, src_interface, dst_interface, src_address, dst_address,
			 protocol, dst_port, action, enabled, priority, description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		p.Name, p.SrcInterface, p.DstInterface,
		p.SrcAddress, p.DstAddress, p.Protocol, p.DstPort,
		p.Action, enabled, p.Priority, p.Description,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

// UpdatePolicy overwrites an existing policy with new field values.
func UpdatePolicy(db *sql.DB, p *Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	enabled := 1
	if !p.Enabled {
		enabled = 0
	}
	_, err := db.Exec(`
		UPDATE policies SET
			name = ?, src_interface = ?, dst_interface = ?,
			src_address = ?, dst_address = ?, protocol = ?, dst_port = ?,
			action = ?, enabled = ?, priority = ?, description = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		p.Name, p.SrcInterface, p.DstInterface,
		p.SrcAddress, p.DstAddress, p.Protocol, p.DstPort,
		p.Action, enabled, p.Priority, p.Description,
		p.ID,
	)
	return err
}

// DeletePolicy removes a policy by id.
func DeletePolicy(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM policies WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("no policy with id %d", id)
	}
	return nil
}
