package data

import (
	"fmt"

	"github.com/Mantelijo/spike-backend/internal/data/cache"
	"github.com/Mantelijo/spike-backend/internal/data/database"
	"github.com/Mantelijo/spike-backend/internal/dto"
)

// TODO DataStore could be an interface and abstract the usage of RedisCache and
// DBService
type DataStore struct {
	RedisCache cache.RedisCache
	DBService  database.DBService
}

// CreateWidget creates a widget in database and sets its initial values in the
// redis cache
func (d *DataStore) CreateWidget(w *dto.Widget) error {
	if err := d.DBService.CreateWidget(w); err != nil {
		return fmt.Errorf("creating widget in db: %w", err)
	}

	if err := d.RedisCache.SetWidget(w); err != nil {
		return fmt.Errorf("setting widget in cache: %w", err)
	}
	return nil
}

func (d *DataStore) RemoveWidget(serialNumber string) error {
	return nil
}

func (d *DataStore) AssociateConnections(incomingWc *dto.WidgetConnections) error {
	// TODO improvements: validate serial numbers - i.e. that they exiss,
	// validate ports, etc. Here we are not doing any of this due to time
	// constraints.

	return d.RedisCache.SetConnections(incomingWc)
}
