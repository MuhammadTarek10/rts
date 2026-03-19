package repository_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/rts/inventory/internal/domain"
	"github.com/rts/inventory/internal/repository"
	"github.com/rts/inventory/internal/testhelpers"
)

func TestStockReceiveFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	pg := testhelpers.StartPostgres(t, ctx)
	defer pg.Cleanup(t, ctx)

	inventoryRepo := repository.NewInventoryRepository(pg.Pool)
	stockRepo := repository.NewStockRepository(pg.Pool)
	movementRepo := repository.NewMovementRepository(pg.Pool)

	// Create an inventory item
	tx, err := inventoryRepo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	item, err := inventoryRepo.Upsert(ctx, tx, &domain.InventoryItem{
		ProductID: "prod-1",
		SKU:       "SKU-TEST-001",
		Title:     "Test Product",
		Status:    "active",
		IsTracked: true,
	})
	if err != nil {
		t.Fatalf("upsert item: %v", err)
	}

	// Create a warehouse
	warehouseRepo := repository.NewWarehouseRepository(pg.Pool)
	wh, err := warehouseRepo.Create(ctx, tx, domain.CreateWarehouseInput{
		Name:      "Test Warehouse",
		Code:      "WH-TEST-01",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	// Create stock level
	sl, err := stockRepo.UpsertWithTx(ctx, tx, item.ID, wh.ID)
	if err != nil {
		t.Fatalf("upsert stock level: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Receive stock
	tx2, err := stockRepo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}

	slForUpdate, err := stockRepo.GetByItemAndWarehouseTx(ctx, tx2, item.ID, wh.ID)
	if err != nil {
		t.Fatalf("get stock for update: %v", err)
	}

	updatedSL, err := stockRepo.UpdateQuantityWithVersion(ctx, tx2, slForUpdate.ID, 100, 0, slForUpdate.Version)
	if err != nil {
		t.Fatalf("update quantity: %v", err)
	}

	performedBy := "test-user"
	_, err = movementRepo.Create(ctx, tx2, &domain.StockMovement{
		InventoryItemID: item.ID,
		WarehouseID:     wh.ID,
		Type:            domain.MovementTypeReceive,
		Quantity:        100,
		PerformedBy:     &performedBy,
		Currency:        "USD",
	})
	if err != nil {
		t.Fatalf("create movement: %v", err)
	}

	if err := tx2.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify stock level
	verifiedSL, err := stockRepo.GetByItemAndWarehouse(ctx, item.ID, wh.ID)
	if err != nil {
		t.Fatalf("verify stock: %v", err)
	}

	if verifiedSL.QuantityOnHand != 100 {
		t.Errorf("expected on_hand 100, got %d", verifiedSL.QuantityOnHand)
	}
	if verifiedSL.QuantityAvailable != 100 {
		t.Errorf("expected available 100, got %d", verifiedSL.QuantityAvailable)
	}
	if verifiedSL.Version != sl.Version+1 {
		t.Errorf("expected version %d, got %d", sl.Version+1, verifiedSL.Version)
	}
	_ = updatedSL // used in commit
}

func TestReserveConfirmFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	pg := testhelpers.StartPostgres(t, ctx)
	defer pg.Cleanup(t, ctx)

	inventoryRepo := repository.NewInventoryRepository(pg.Pool)
	stockRepo := repository.NewStockRepository(pg.Pool)
	reservationRepo := repository.NewReservationRepository(pg.Pool)

	// Setup: item + warehouse + stock
	tx, err := inventoryRepo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	item, err := inventoryRepo.Upsert(ctx, tx, &domain.InventoryItem{
		ProductID: "prod-2",
		SKU:       "SKU-TEST-002",
		Title:     "Reserve Test Product",
		Status:    "active",
		IsTracked: true,
	})
	if err != nil {
		t.Fatalf("upsert item: %v", err)
	}

	warehouseRepo := repository.NewWarehouseRepository(pg.Pool)
	wh, err := warehouseRepo.Create(ctx, tx, domain.CreateWarehouseInput{
		Name:      "Reserve Warehouse",
		Code:      "WH-RES-01",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	_, err = stockRepo.UpsertWithTx(ctx, tx, item.ID, wh.ID)
	if err != nil {
		t.Fatalf("upsert stock level: %v", err)
	}

	// Set stock to 50
	sl, _ := stockRepo.GetByItemAndWarehouseTx(ctx, tx, item.ID, wh.ID)
	_, err = stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, 50, 0, sl.Version)
	if err != nil {
		t.Fatalf("set stock: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Reserve 10 units
	tx2, err := reservationRepo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}

	sl2, _ := stockRepo.GetByItemAndWarehouseTx(ctx, tx2, item.ID, wh.ID)
	_, err = stockRepo.UpdateQuantityWithVersion(ctx, tx2, sl2.ID, sl2.QuantityOnHand, sl2.QuantityReserved+10, sl2.Version)
	if err != nil {
		t.Fatalf("reserve stock: %v", err)
	}

	res, err := reservationRepo.Create(ctx, tx2, &domain.Reservation{
		InventoryItemID: item.ID,
		WarehouseID:     wh.ID,
		OrderID:         "order-001",
		Quantity:        10,
		Status:          domain.ReservationStatusActive,
		ExpiresAt:       time.Now().Add(15 * time.Minute),
	})
	if err != nil {
		t.Fatalf("create reservation: %v", err)
	}

	if err := tx2.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify reservation
	if res.Status != domain.ReservationStatusActive {
		t.Errorf("expected active status, got %s", res.Status)
	}

	// Verify stock
	verifiedSL, _ := stockRepo.GetByItemAndWarehouse(ctx, item.ID, wh.ID)
	if verifiedSL.QuantityReserved != 10 {
		t.Errorf("expected reserved 10, got %d", verifiedSL.QuantityReserved)
	}
	if verifiedSL.QuantityAvailable != 40 {
		t.Errorf("expected available 40, got %d", verifiedSL.QuantityAvailable)
	}

	// Confirm reservation
	tx3, err := reservationRepo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx3: %v", err)
	}

	err = reservationRepo.Confirm(ctx, tx3, res.ID)
	if err != nil {
		t.Fatalf("confirm reservation: %v", err)
	}

	if err := tx3.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Verify confirmed reservations
	reservations, err := reservationRepo.GetByOrderID(ctx, "order-001")
	if err != nil {
		t.Fatalf("get reservations: %v", err)
	}
	if len(reservations) != 1 {
		t.Fatalf("expected 1 reservation, got %d", len(reservations))
	}
	if reservations[0].Status != domain.ReservationStatusConfirmed {
		t.Errorf("expected confirmed status, got %s", reservations[0].Status)
	}
}

func TestConcurrentReservations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	pg := testhelpers.StartPostgres(t, ctx)
	defer pg.Cleanup(t, ctx)

	inventoryRepo := repository.NewInventoryRepository(pg.Pool)
	stockRepo := repository.NewStockRepository(pg.Pool)
	reservationRepo := repository.NewReservationRepository(pg.Pool)

	// Setup: item + warehouse + stock of 10
	tx, err := inventoryRepo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	item, err := inventoryRepo.Upsert(ctx, tx, &domain.InventoryItem{
		ProductID: "prod-3",
		SKU:       "SKU-TEST-003",
		Title:     "Concurrent Test Product",
		Status:    "active",
		IsTracked: true,
	})
	if err != nil {
		t.Fatalf("upsert item: %v", err)
	}

	warehouseRepo := repository.NewWarehouseRepository(pg.Pool)
	wh, err := warehouseRepo.Create(ctx, tx, domain.CreateWarehouseInput{
		Name:      "Concurrent Warehouse",
		Code:      "WH-CON-01",
		IsDefault: true,
	})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}

	_, err = stockRepo.UpsertWithTx(ctx, tx, item.ID, wh.ID)
	if err != nil {
		t.Fatalf("upsert stock level: %v", err)
	}

	sl, _ := stockRepo.GetByItemAndWarehouseTx(ctx, tx, item.ID, wh.ID)
	_, err = stockRepo.UpdateQuantityWithVersion(ctx, tx, sl.ID, 10, 0, sl.Version)
	if err != nil {
		t.Fatalf("set stock: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Try 5 concurrent reservations of 3 units each (only 3 should succeed: 3*3=9 <= 10, 4*3=12 > 10)
	concurrency := 5
	reserveQty := 3
	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(orderNum int) {
			defer wg.Done()

			rtx, err := reservationRepo.BeginTx(ctx)
			if err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			rsl, err := stockRepo.GetByItemAndWarehouseTx(ctx, rtx, item.ID, wh.ID)
			if err != nil || rsl == nil || rsl.QuantityAvailable < reserveQty {
				_ = rtx.Rollback(ctx)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			newReserved := rsl.QuantityReserved + reserveQty
			_, err = stockRepo.UpdateQuantityWithVersion(ctx, rtx, rsl.ID, rsl.QuantityOnHand, newReserved, rsl.Version)
			if err != nil {
				_ = rtx.Rollback(ctx)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			_, err = reservationRepo.Create(ctx, rtx, &domain.Reservation{
				InventoryItemID: item.ID,
				WarehouseID:     wh.ID,
				OrderID:         fmt.Sprintf("order-con-%d", orderNum),
				Quantity:        reserveQty,
				Status:          domain.ReservationStatusActive,
				ExpiresAt:       time.Now().Add(15 * time.Minute),
			})
			if err != nil {
				_ = rtx.Rollback(ctx)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			if err := rtx.Commit(ctx); err != nil {
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	// Due to optimistic locking, not all should succeed
	t.Logf("concurrent reservations: %d succeeded, %d failed", successCount, failCount)

	// Verify final stock state is consistent
	finalSL, err := stockRepo.GetByItemAndWarehouse(ctx, item.ID, wh.ID)
	if err != nil {
		t.Fatalf("get final stock: %v", err)
	}

	expectedReserved := successCount * reserveQty
	if finalSL.QuantityReserved != expectedReserved {
		t.Errorf("expected reserved %d, got %d", expectedReserved, finalSL.QuantityReserved)
	}
	if finalSL.QuantityOnHand != 10 {
		t.Errorf("expected on_hand 10, got %d", finalSL.QuantityOnHand)
	}
	if finalSL.QuantityAvailable != 10-expectedReserved {
		t.Errorf("expected available %d, got %d", 10-expectedReserved, finalSL.QuantityAvailable)
	}
}
