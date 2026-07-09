package service

import (
	"errors"
	"time"

	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
	"wtr-lab-clone/backend/internal/ticket"
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
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return 0, err
	}
	return user.Tickets, nil
}

func (s *TicketService) Spend(userID uint, amount float64, refType, note string) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		if user.Tickets < amount {
			return ErrInsufficientTickets
		}

		if err := tx.Model(&user).Update("tickets", user.Tickets-amount).Error; err != nil {
			return err
		}

		return tx.Create(&model.TicketTransaction{
			UserID: userID,
			Amount: -amount,
			Type:   "spend",
			RefType:  refType,
			Date:   time.Now(),
			Note:   note,
		}).Error
	})
}

func (s *TicketService) Award(userID uint, amount float64, refType, note string) error {
	if amount <= 0 {
		return errors.New("invalid amount")
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		if err := tx.Model(&user).Update("tickets", user.Tickets+amount).Error; err != nil {
			return err
		}

		return tx.Create(&model.TicketTransaction{
			UserID: userID,
			Amount: amount,
			Type:   "reward",
			RefType:  refType,
			Date:   time.Now(),
			Note:   note,
		}).Error
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
		if err := tx.Model(&user).Updates(map[string]interface{}{
			"tickets":          user.Tickets + reward,
			"last_daily_claim": &now,
		}).Error; err != nil {
			return err
		}
		return tx.Create(&model.TicketTransaction{
			UserID: userID,
			Amount: reward,
			Type:   "reward",
			RefType: "daily",
			Date:   now,
		}).Error
	})

	return reward, err
}

func todayMakassarBoundary() time.Time {
	loc := time.FixedZone("WITA", 8*3600)
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}
