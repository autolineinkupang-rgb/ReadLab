package ticket

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
)

func setupTicketTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	db.AutoMigrate(&model.TicketConfig{})
	return db
}

func TestNewConfig_LoadsExistingConfigs(t *testing.T) {
	db := setupTicketTestDB(t)
	db.Create(&model.TicketConfig{Key: "daily_reward", Value: 5, Label: "Daily Reward"})
	db.Create(&model.TicketConfig{Key: "edit_reset_cost", Value: 20, Label: "Edit Reset Cost"})

	cfg := NewConfig(db)

	if v := cfg.Get("daily_reward"); v != 5 {
		t.Errorf("expected daily_reward=5, got %v", v)
	}
	if v := cfg.Get("edit_reset_cost"); v != 20 {
		t.Errorf("expected edit_reset_cost=20, got %v", v)
	}
}

func TestNewConfig_EmptyDatabase(t *testing.T) {
	db := setupTicketTestDB(t)
	cfg := NewConfig(db)

	// Should not panic, cache should be empty
	if v := cfg.Get("daily_reward"); v != 0 {
		t.Errorf("expected 0 for missing key, got %v", v)
	}
}

func TestConfigGet_ExistingKey(t *testing.T) {
	db := setupTicketTestDB(t)
	db.Create(&model.TicketConfig{Key: "daily_reward", Value: 3, Label: "Daily"})
	cfg := NewConfig(db)

	v := cfg.Get("daily_reward")
	if v != 3 {
		t.Errorf("expected 3, got %v", v)
	}
}

func TestConfigGet_MissingKey(t *testing.T) {
	db := setupTicketTestDB(t)
	db.Create(&model.TicketConfig{Key: "daily_reward", Value: 3, Label: "Daily"})
	cfg := NewConfig(db)

	v := cfg.Get("nonexistent_key")
	if v != 0 {
		t.Errorf("expected 0 for missing key, got %v", v)
	}
}

func TestConfigGet_ZeroValue(t *testing.T) {
	db := setupTicketTestDB(t)
	db.Create(&model.TicketConfig{Key: "zero_config", Value: 0, Label: "Zero"})
	cfg := NewConfig(db)

	// A key that exists with value 0 should return 0, same as missing
	// This is a known limitation of the design — 0 is used as the default
	v := cfg.Get("zero_config")
	if v != 0 {
		t.Errorf("expected 0, got %v", v)
	}
}

func TestConfigList_ReturnsAll(t *testing.T) {
	db := setupTicketTestDB(t)
	db.Create(&model.TicketConfig{Key: "beta", Value: 2, Label: "B"})
	db.Create(&model.TicketConfig{Key: "alpha", Value: 1, Label: "A"})
	db.Create(&model.TicketConfig{Key: "gamma", Value: 3, Label: "G"})

	cfg := NewConfig(db)
	configs := cfg.List()

	if len(configs) != 3 {
		t.Fatalf("expected 3 configs, got %d", len(configs))
	}
	// Should be sorted by key ASC
	if configs[0].Key != "alpha" {
		t.Errorf("expected first key 'alpha', got '%s'", configs[0].Key)
	}
	if configs[1].Key != "beta" {
		t.Errorf("expected second key 'beta', got '%s'", configs[1].Key)
	}
	if configs[2].Key != "gamma" {
		t.Errorf("expected third key 'gamma', got '%s'", configs[2].Key)
	}
}

func TestConfigList_EmptyDatabase(t *testing.T) {
	db := setupTicketTestDB(t)
	cfg := NewConfig(db)

	configs := cfg.List()
	if len(configs) != 0 {
		t.Errorf("expected empty list, got %d items", len(configs))
	}
}

func TestConfigUpdate_ExistingKey(t *testing.T) {
	db := setupTicketTestDB(t)
	db.Create(&model.TicketConfig{Key: "daily_reward", Value: 2, Label: "Daily"})

	cfg := NewConfig(db)
	if v := cfg.Get("daily_reward"); v != 2 {
		t.Fatalf("precondition: expected 2, got %v", v)
	}

	err := cfg.Update("daily_reward", 10)
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	// Cache should be refreshed
	if v := cfg.Get("daily_reward"); v != 10 {
		t.Errorf("expected daily_reward=10 after update, got %v", v)
	}

	// DB should also be updated
	var tc model.TicketConfig
	db.Where("key = ?", "daily_reward").First(&tc)
	if tc.Value != 10 {
		t.Errorf("expected DB value=10, got %v", tc.Value)
	}
}

func TestConfigUpdate_NonexistentKey(t *testing.T) {
	db := setupTicketTestDB(t)
	cfg := NewConfig(db)

	// Updating a key that doesn't exist won't match any rows but shouldn't error
	err := cfg.Update("nonexistent", 42)
	if err != nil {
		t.Fatalf("Update on nonexistent key should not error, got: %v", err)
	}

	// Key should still not exist in cache
	if v := cfg.Get("nonexistent"); v != 0 {
		t.Errorf("expected 0 for nonexistent key, got %v", v)
	}
}

func TestConfigUpdate_ReloadsCache(t *testing.T) {
	db := setupTicketTestDB(t)
	db.Create(&model.TicketConfig{Key: "cost_a", Value: 10, Label: "Cost A"})
	db.Create(&model.TicketConfig{Key: "cost_b", Value: 20, Label: "Cost B"})

	cfg := NewConfig(db)

	// Update one key, verify the other is still correct
	err := cfg.Update("cost_a", 50)
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}

	if v := cfg.Get("cost_a"); v != 50 {
		t.Errorf("expected cost_a=50, got %v", v)
	}
	if v := cfg.Get("cost_b"); v != 20 {
		t.Errorf("expected cost_b=20 (unchanged), got %v", v)
	}
}