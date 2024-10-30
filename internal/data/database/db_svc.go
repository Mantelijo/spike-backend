package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/Mantelijo/spike-backend/internal/dto"
	"github.com/jackc/pgx/v5"
)

type DBService interface {
	// CreateWidget creates a new widget in the database. Returns the inserted
	// widget with the new ID.
	CreateWidget(w *dto.Widget) error

	// DeleteWidget deletes the widget records from database
	DeleteWidget(serialNumber string) error

	// UpdateAssociations updates the connection associations
	UpdateAssociations([]*dto.WidgetConnections) error
}

func NewDbService(dbDsn string) (DBService, error) {
	conn, err := pgx.Connect(context.Background(), dbDsn)
	if err != nil {
		return nil, err
	}

	return &dbService{
		conn: conn,
	}, nil
}

type dbService struct {
	conn *pgx.Conn
}

func (d *dbService) CreateWidget(w *dto.Widget) error {
	tx, err := d.conn.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())
	sql1 := `
		insert into widgets (name, serial_number, ports_bitmask)
		values ($1, $2, $3);
	`
	if _, err := tx.Exec(context.Background(), sql1, w.Name, w.SerialNumber, w.PortBitmap.ToBitString()); err != nil {
		return err
	}

	sql2 := `insert into widget_connections(widget_sn) values ($1);`
	if _, err := tx.Exec(context.Background(), sql2, w.SerialNumber); err != nil {
		return err
	}
	return tx.Commit(context.Background())
}

func (d *dbService) DeleteWidget(serialNumber string) error {
	panic("DeleteWidget is not implemented")
	return nil
}

func (d *dbService) UpdateAssociations(connections []*dto.WidgetConnections) error {
	if len(connections) == 0 {
		return nil // No connections to upsert
	}

	// Start building the SQL statement
	sql := `INSERT INTO widget_connections (widget_sn, p_peer_sn, r_peer_sn, q_peer_sn) VALUES `
	values := []interface{}{}

	// Populate the SQL statement with the connection data
	for i, conn := range connections {
		// Calculate the offset for the current connection
		offset := i * 4 // 4 fields to insert

		// Append the values for the current connection
		sql += fmt.Sprintf("($%d, $%d, $%d, $%d),", offset+1, offset+2, offset+3, offset+4)
		var pPeer any = nil
		if conn.P_PeerSerialNumber != "" {
			pPeer = conn.P_PeerSerialNumber
		}
		var rPeer any = nil
		if conn.R_PeerSerialNumber != "" {
			rPeer = conn.R_PeerSerialNumber
		}
		var qPeer any = nil
		if conn.Q_PeerSerialNumber != "" {
			qPeer = conn.Q_PeerSerialNumber
		}
		values = append(values, conn.SerialNumber, pPeer, rPeer, qPeer)
	}

	// Remove the trailing comma
	sql = strings.TrimSuffix(sql, ",")

	// Add the ON CONFLICT clause for upserting
	sql += ` ON CONFLICT (widget_sn) DO UPDATE SET 
                p_peer_sn = EXCLUDED.p_peer_sn,
                r_peer_sn = EXCLUDED.r_peer_sn,
                q_peer_sn = EXCLUDED.q_peer_sn;`

	// Execute the query
	_, err := d.conn.Exec(context.Background(), sql, values...)
	return err
}
