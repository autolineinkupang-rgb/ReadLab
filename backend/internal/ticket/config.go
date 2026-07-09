package ticket

import (
	"sync"
	"time"

	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type Config struct {
	db    *gorm.DB
	mu    sync.RWMutex
	cache map[string]float64
}

func NewConfig(db *gorm.DB) *Config {
	c := &Config{
		db:    db,
		cache: make(map[string]float64),
	}
	c.Reload()
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			c.Reload()
		}
	}()
	return c
}

func (c *Config) Reload() {
	var configs []model.TicketConfig
	if err := c.db.Find(&configs).Error; err != nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cfg := range configs {
		c.cache[cfg.Key] = cfg.Value
	}
}

func (c *Config) Get(key string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.cache[key]; ok {
		return v
	}
	return 0
}

func (c *Config) List() []model.TicketConfig {
	var configs []model.TicketConfig
	c.db.Order("key ASC").Find(&configs)
	return configs
}

func (c *Config) Update(key string, value float64) error {
	if err := c.db.Model(&model.TicketConfig{}).Where("key = ?", key).Update("value", value).Error; err != nil {
		return err
	}
	c.Reload()
	return nil
}

func (c *Config) Spend(db *gorm.DB, userID uint, cost float64, refType string, refID uint, note string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}
		if user.Tickets < cost {
			return ErrInsufficientTickets
		}
		if err := tx.Model(&user).Update("tickets", gorm.Expr("tickets - ?", cost)).Error; err != nil {
			return err
		}
		return tx.Create(&model.TicketTransaction{
			UserID:  userID,
			Amount:  -cost,
			Type:    "spend",
			RefType: refType,
			RefID:   refID,
			Note:    note,
		}).Error
	})
}

func (c *Config) Award(db *gorm.DB, userID uint, amount float64, refType, note string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).Where("id = ?", userID).
			Update("tickets", gorm.Expr("tickets + ?", amount)).Error; err != nil {
			return err
		}
		return tx.Create(&model.TicketTransaction{
			UserID:  userID,
			Amount:  amount,
			Type:    "reward",
			RefType: refType,
			Note:    note,
		}).Error
	})
}
