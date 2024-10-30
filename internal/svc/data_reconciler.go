package svc

import (
	"log/slog"
	"time"

	"github.com/Mantelijo/spike-backend/internal/data/cache"
	"github.com/Mantelijo/spike-backend/internal/data/database"
)

// TODO extract to a separate service, worker pool, etc.

// RunDataReconciler fetches recently updated widgets from the cache and updates
// the database with the new connections. This func blocks.
func RunDataReconciler(rc cache.RedisCache, db database.DBService) {

	for {
		t1 := time.Now()
		connUpdates, err := rc.RetrieveRecentUpdates(1000)
		if err != nil {
			slog.Error("Failed to retrieve recent updates", slog.Any("error", err))
			time.Sleep(time.Second)
			continue
		}

		if len(connUpdates) == 0 {
			slog.Debug("No recent updates")
			time.Sleep(time.Second)
			continue
		}

		// Process the updates
		if err := db.UpdateAssociations(connUpdates); err != nil {
			slog.Error("Failed to update associations", slog.Any("error", err))
			time.Sleep(time.Second)
			continue
		}

		end := time.Since(t1)
		slog.Info("Data reconciler run completed", slog.Any("duration", end))
	}

}
