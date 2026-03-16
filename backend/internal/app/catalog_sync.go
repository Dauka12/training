package app

import (
	"context"
	"fmt"
	"time"
)

func (a *App) RunCatalogSyncOnce(ctx context.Context, limit int) error {
	if a.catalogImporter == nil {
		return fmt.Errorf("catalog importer unavailable")
	}
	equipment, exercises, err := a.catalogImporter.ImportWger(ctx, limit)
	if err != nil {
		return err
	}

	a.mu.Lock()
	mergeImportedEquipment(&a.equipmentCatalog, equipment)
	mergeImportedExercises(&a.exerciseCatalog, exercises)
	a.recordAuditLocked("system", "sync_catalog_wger", "catalog", "wger")
	a.persistStateLocked()
	a.mu.Unlock()

	return nil
}

func (a *App) StartCatalogSync(ctx context.Context, interval time.Duration, limit int) {
	if interval <= 0 || a.catalogImporter == nil {
		return
	}

	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := a.RunCatalogSyncOnce(ctx, limit); err != nil {
					a.log.Error("catalog sync failed", "error", err)
				}
			}
		}
	}()
}
