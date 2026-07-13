package service

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/ticket"
)

var (
	ErrInsufficientTickets = errors.New("insufficient tickets")
)

type TicketService struct {
	DB     *gorm.DB
	Config *ticket.Config
}

func NewTicketService(db *gorm.DB, cfg *ticket.Config) *TicketService {
	return &TicketService{DB: db, Config: cfg}
}

func (s *TicketService) GetBalance(userID uint) (float64, error) {
	var sum float64
	err := s.DB.Model(&model.TicketUnit{}).
		Where("user_id = ? AND status = 'active'", userID).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

func (s *TicketService) Spend(userID uint, amount float64, refType, note string) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var sum float64
		tx.Model(&model.TicketUnit{}).
			Where("user_id = ? AND status = 'active'", userID).
			Select("COALESCE(SUM(amount), 0)").Scan(&sum)

		if sum < amount {
			return ErrInsufficientTickets
		}

		var units []model.TicketUnit
		tx.Where("user_id = ? AND status = 'active'", userID).
			Order("created_at ASC, id ASC").
			Find(&units)

		remaining := amount
		now := time.Now()
		for _, unit := range units {
			if remaining <= 0 {
				break
			}
			if unit.Amount <= remaining {
				tx.Model(&unit).Updates(map[string]interface{}{
					"status":   "banked",
					"spent_at": &now,
				})
				remaining -= unit.Amount
			} else {
				excess := unit.Amount - remaining
				tx.Model(&unit).Updates(map[string]interface{}{
					"status":   "banked",
					"spent_at": &now,
				})
				tx.Create(&model.TicketUnit{
					Serial: model.NewSerial(),
					UserID: userID,
					Amount: excess,
					Status: "active",
				})
				remaining = 0
			}
		}

		tx.Create(&model.TicketTransaction{
			UserID:  userID,
			Amount:  -amount,
			Type:    "spend",
			RefType: refType,
			Date:    now,
			Note:    note,
		})

		var newSum float64
		tx.Model(&model.TicketUnit{}).
			Where("user_id = ? AND status = 'active'", userID).
			Select("COALESCE(SUM(amount), 0)").Scan(&newSum)
		tx.Model(&model.User{}).Where("id = ?", userID).Update("tickets", newSum)

		return nil
	})
}

func (s *TicketService) Award(userID uint, amount float64, refType, note string) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		tx.Create(&model.TicketUnit{
			Serial: model.NewSerial(),
			UserID: userID,
			Amount: amount,
			Status: "active",
		})

		tx.Create(&model.TicketTransaction{
			UserID:  userID,
			Amount:  amount,
			Type:    "reward",
			RefType: refType,
			Date:    time.Now(),
			Note:    note,
		})

		var sum float64
		tx.Model(&model.TicketUnit{}).
			Where("user_id = ? AND status = 'active'", userID).
			Select("COALESCE(SUM(amount), 0)").Scan(&sum)
		tx.Model(&model.User{}).Where("id = ?", userID).Update("tickets", sum)

		return nil
	})
}

func (s *TicketService) GetTransactions(userID uint, page, limit int) ([]model.TicketTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	s.DB.Model(&model.TicketTransaction{}).Where("user_id = ?", userID).Count(&total)

	var txns []model.TicketTransaction
	offset := (page - 1) * limit
	err := s.DB.Where("user_id = ?", userID).Order("date DESC").Offset(offset).Limit(limit).Find(&txns).Error

	return txns, total, err
}

func (s *TicketService) DailyRewardEligible(userID uint) (bool, float64, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return false, 0, err
	}

	todayStart := todayMakassarBoundary()
	canClaim := user.LastDailyClaim == nil || user.LastDailyClaim.Before(todayStart)
	reward := s.Config.Get("daily_reward")
	return canClaim, reward, nil
}

func (s *TicketService) ClaimDailyReward(userID uint) (float64, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return 0, err
	}

	todayStart := todayMakassarBoundary()
	if user.LastDailyClaim != nil && !user.LastDailyClaim.Before(todayStart) {
		return 0, errors.New("daily reward already claimed today")
	}

	reward := s.Config.Get("daily_reward")
	if reward <= 0 {
		reward = 2
	}

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		tx.Create(&model.TicketUnit{
			Serial: model.NewSerial(),
			UserID: userID,
			Amount: reward,
			Status: "active",
		})
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"last_daily_claim": &now,
		}).Error; err != nil {
			return err
		}
		tx.Create(&model.TicketTransaction{
			UserID: userID,
			Amount: reward,
			Type:   "reward",
			RefType: "daily",
			Date:   now,
		})
		var sum float64
		tx.Model(&model.TicketUnit{}).
			Where("user_id = ? AND status = 'active'", userID).
			Select("COALESCE(SUM(amount), 0)").Scan(&sum)
		tx.Model(&model.User{}).Where("id = ?", userID).Update("tickets", sum)
		return nil
	})

	return reward, err
}

func todayMakassarBoundary() time.Time {
	loc := time.FixedZone("WITA", 8*3600)
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}
