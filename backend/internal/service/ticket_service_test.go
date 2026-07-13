package service

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/ticket"
)

func setupTicketServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(
		&model.User{}, &model.TicketConfig{},
		&model.TicketUnit{}, &model.TicketTransaction{},
	)
	return db
}

func createTestTicketUser(t *testing.T, db *gorm.DB) model.User {
	user := model.User{
		Username:     fmt.Sprintf("ticketuser%d", time.Now().UnixNano()),
		Email:        fmt.Sprintf("ticket%d@example.com", time.Now().UnixNano()),
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxABCDEFGHIJ",
		DisplayName:  "Ticket User",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func setupTicketService(t *testing.T) (*TicketService, *gorm.DB) {
	db := setupTicketServiceTestDB(t)
	// Seed config
	db.Create(&model.TicketConfig{Key: "daily_reward", Value: 3, Label: "Daily Reward"})
	cfg := ticket.NewConfig(db)
	svc := NewTicketService(db, cfg)
	return svc, db
}

func TestGetBalance_NoTickets(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	balance, err := svc.GetBalance(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 0 {
		t.Errorf("expected balance 0, got %v", balance)
	}
}

func TestGetBalance_WithActiveUnits(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// Create some active ticket units
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 10, Status: "active"})
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 5, Status: "active"})

	balance, err := svc.GetBalance(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 15 {
		t.Errorf("expected balance 15, got %v", balance)
	}
}

func TestGetBalance_IgnoresBankedUnits(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 10, Status: "active"})
	now := time.Now()
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 5, Status: "banked", SpentAt: &now})

	balance, err := svc.GetBalance(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 10 {
		t.Errorf("expected balance 10 (only active), got %v", balance)
	}
}

func TestSpend_InsufficientTickets(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// Only 5 tickets, trying to spend 10
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 5, Status: "active"})

	err := svc.Spend(user.ID, 10, "test", "test spend")
	if err == nil {
		t.Fatal("expected error for insufficient tickets")
	}
	if err.Error() != "insufficient tickets" {
		t.Errorf("expected 'insufficient tickets', got '%s'", err.Error())
	}
}

func TestSpend_InvalidAmount(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	tests := []struct {
		name   string
		amount float64
	}{
		{"zero amount", 0},
		{"negative amount", -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Spend(user.ID, tt.amount, "test", "test")
			if err == nil {
				t.Error("expected error for invalid amount")
			}
		})
	}
}

func TestSpend_ExactAmount(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// 10 tickets, spend exactly 10
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 10, Status: "active"})

	err := svc.Spend(user.ID, 10, "purchase", "bought something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check balance
	balance, _ := svc.GetBalance(user.ID)
	if balance != 0 {
		t.Errorf("expected balance 0, got %v", balance)
	}

	// Check transaction was created
	var count int64
	db.Model(&model.TicketTransaction{}).Where("user_id = ? AND type = ? AND amount = ?", user.ID, "spend", -10).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 spend transaction, got %d", count)
	}

	// Unit should be banked
	var units []model.TicketUnit
	db.Where("user_id = ?", user.ID).Find(&units)
	for _, u := range units {
		if u.Status != "banked" {
			t.Errorf("expected all units to be banked, got status '%s'", u.Status)
		}
	}
}

func TestSpend_PartialSpend(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// Two units: 10 and 5 = 15 total, spend 12
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 10, Status: "active"})
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 5, Status: "active"})

	err := svc.Spend(user.ID, 12, "purchase", "partial spend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Balance should be 3 (15 - 12)
	balance, _ := svc.GetBalance(user.ID)
	if balance != 3 {
		t.Errorf("expected balance 3, got %v", balance)
	}

	// Verify: first unit fully consumed, second partially consumed
	// First unit (10) → fully banked
	// Second unit (5) → banked, new active unit of 3 created
	var activeUnits []model.TicketUnit
	db.Where("user_id = ? AND status = 'active'", user.ID).Find(&activeUnits)
	if len(activeUnits) != 1 {
		t.Fatalf("expected 1 active unit (remainder), got %d", len(activeUnits))
	}
	if activeUnits[0].Amount != 3 {
		t.Errorf("expected remainder unit of 3, got %v", activeUnits[0].Amount)
	}
}

func TestSpend_SplitsSingleUnit(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// Single unit of 10, spend 3 → 7 should remain
	db.Create(&model.TicketUnit{Serial: model.NewSerial(), UserID: user.ID, Amount: 10, Status: "active"})

	err := svc.Spend(user.ID, 3, "test", "split unit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	balance, _ := svc.GetBalance(user.ID)
	if balance != 7 {
		t.Errorf("expected balance 7, got %v", balance)
	}

	// Check transaction
	var txn model.TicketTransaction
	db.Where("user_id = ? AND type = ?", user.ID, "spend").First(&txn)
	if txn.Amount != -3 {
		t.Errorf("expected transaction amount -3, got %v", txn.Amount)
	}
}

func TestAward_CreatesUnitAndTransaction(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	err := svc.Award(user.ID, 25, "daily", "daily reward")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check balance
	balance, _ := svc.GetBalance(user.ID)
	if balance != 25 {
		t.Errorf("expected balance 25, got %v", balance)
	}

	// Check ticket unit was created
	var units []model.TicketUnit
	db.Where("user_id = ? AND status = 'active'", user.ID).Find(&units)
	if len(units) != 1 {
		t.Fatalf("expected 1 active unit, got %d", len(units))
	}
	if units[0].Amount != 25 {
		t.Errorf("expected unit amount 25, got %v", units[0].Amount)
	}

	// Check transaction was created
	var txn model.TicketTransaction
	db.Where("user_id = ? AND type = ?", user.ID, "reward").First(&txn)
	if txn.Amount != 25 {
		t.Errorf("expected transaction amount 25, got %v", txn.Amount)
	}
	if txn.RefType != "daily" {
		t.Errorf("expected ref_type 'daily', got '%s'", txn.RefType)
	}
}

func TestAward_InvalidAmount(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	err := svc.Award(user.ID, 0, "daily", "zero award")
	if err == nil {
		t.Error("expected error for zero amount")
	}

	err = svc.Award(user.ID, -5, "daily", "negative award")
	if err == nil {
		t.Error("expected error for negative amount")
	}
}

func TestAward_MultipleAwardsAccumulate(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	if err := svc.Award(user.ID, 10, "daily", "day 1"); err != nil {
		t.Fatalf("first award failed: %v", err)
	}
	if err := svc.Award(user.ID, 15, "contribution", "novel contribution"); err != nil {
		t.Fatalf("second award failed: %v", err)
	}

	balance, _ := svc.GetBalance(user.ID)
	if balance != 25 {
		t.Errorf("expected balance 25, got %v", balance)
	}

	var txnCount int64
	db.Model(&model.TicketTransaction{}).Where("user_id = ?", user.ID).Count(&txnCount)
	if txnCount != 2 {
		t.Errorf("expected 2 transactions, got %d", txnCount)
	}
}

func TestGetTransactions_Pagination(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// Create 5 transactions
	for i := 0; i < 5; i++ {
		db.Create(&model.TicketTransaction{
			UserID:  user.ID,
			Amount:  float64(i + 1),
			Type:    "reward",
			RefType: "daily",
			Date:    time.Now().Add(time.Duration(i) * time.Hour),
		})
	}

	// Page 1, limit 2
	txns, total, err := svc.GetTransactions(user.ID, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if len(txns) != 2 {
		t.Fatalf("expected 2 txns on page 1, got %d", len(txns))
	}

	// Should be ordered by date DESC (most recent first)
	if txns[0].Amount != 5 {
		t.Errorf("expected first txn amount 5 (most recent), got %v", txns[0].Amount)
	}

	// Page 2, limit 2
	txns2, _, err := svc.GetTransactions(user.ID, 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns2) != 2 {
		t.Fatalf("expected 2 txns on page 2, got %d", len(txns2))
	}

	// Page 3, limit 2 (only 1 left)
	txns3, _, err := svc.GetTransactions(user.ID, 3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns3) != 1 {
		t.Fatalf("expected 1 txn on page 3, got %d", len(txns3))
	}
}

func TestGetTransactions_ClampsLimit(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// limit > 100 should be clamped to 20
	txns, _, err := svc.GetTransactions(user.ID, 1, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txns) != 0 {
		t.Errorf("expected 0 txns, got %d", len(txns))
	}
}

func TestGetTransactions_ClampsPage(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// page 0 should be treated as page 1
	txns, _, err := svc.GetTransactions(user.ID, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txns == nil {
		t.Error("expected non-nil result even with page 0")
	}
}

func TestDailyRewardEligible_NeverClaimed(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	eligible, reward, err := svc.DailyRewardEligible(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eligible {
		t.Error("expected eligible for never-claimed user")
	}
	if reward != 3 {
		t.Errorf("expected reward 3 (from config), got %v", reward)
	}
}

func TestDailyRewardEligible_AlreadyClaimed(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// Set LastDailyClaim to now
	now := time.Now()
	db.Model(&user).Update("last_daily_claim", now)

	eligible, _, err := svc.DailyRewardEligible(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if eligible {
		t.Error("expected not eligible when already claimed today")
	}
}

func TestDailyRewardEligible_ClaimedYesterday(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// Set LastDailyClaim to 25 hours ago (yesterday in Makassar)
	yesterday := time.Now().Add(-25 * time.Hour)
	db.Model(&user).Update("last_daily_claim", yesterday)

	eligible, _, err := svc.DailyRewardEligible(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !eligible {
		t.Error("expected eligible when last claim was yesterday")
	}
}

func TestClaimDailyReward_Success(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	reward, err := svc.ClaimDailyReward(user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reward != 3 {
		t.Errorf("expected reward 3, got %v", reward)
	}

	// Check user was credited
	var refreshed model.User
	db.First(&refreshed, user.ID)
	if refreshed.Tickets != 3 {
		t.Errorf("expected user tickets=3, got %v", refreshed.Tickets)
	}
	if refreshed.LastDailyClaim == nil {
		t.Error("expected LastDailyClaim to be set")
	}

	// Check balance
	balance, _ := svc.GetBalance(user.ID)
	if balance != 3 {
		t.Errorf("expected balance 3, got %v", balance)
	}
}

func TestClaimDailyReward_DuplicateClaim(t *testing.T) {
	svc, db := setupTicketService(t)
	user := createTestTicketUser(t, db)

	// First claim should succeed
	_, err := svc.ClaimDailyReward(user.ID)
	if err != nil {
		t.Fatalf("first claim failed: %v", err)
	}

	// Second claim should fail
	_, err = svc.ClaimDailyReward(user.ID)
	if err == nil {
		t.Fatal("expected error for duplicate daily claim")
	}
	if err.Error() != "daily reward already claimed today" {
		t.Errorf("expected 'daily reward already claimed today', got '%s'", err.Error())
	}
}

func TestClaimDailyReward_NonexistentUser(t *testing.T) {
	svc, _ := setupTicketService(t)

	_, err := svc.ClaimDailyReward(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}

func TestDailyRewardEligible_NonexistentUser(t *testing.T) {
	svc, _ := setupTicketService(t)

	_, _, err := svc.DailyRewardEligible(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}