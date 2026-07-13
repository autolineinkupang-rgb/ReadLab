package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"readlab/backend/internal/model"
	"readlab/backend/internal/service"
	"readlab/backend/internal/ticket"
)

type RewardHandler struct {
	DB        *gorm.DB
	Config    *ticket.Config
	TicketSvc *service.TicketService
}

func NewRewardHandler(db *gorm.DB, cfg *ticket.Config, ticketSvc *service.TicketService) *RewardHandler {
	return &RewardHandler{DB: db, Config: cfg, TicketSvc: ticketSvc}
}

func makassarNow() time.Time {
	return ticket.MakassarNow()
}

func todayMakassarBoundary() time.Time {
	return ticket.TodayMakassarBoundary()
}

func (h *RewardHandler) awardTickets(userID uint, amount float64, refType, note string) error {
	return h.Config.Award(h.DB, userID, amount, refType, note)
}

func (h *RewardHandler) ClaimDaily(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	uid, ok := userID.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}
	var user model.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	todayStart := todayMakassarBoundary()
	now := makassarNow()
	reward := h.Config.Get("daily_reward")

	if user.LastDailyClaim != nil && !user.LastDailyClaim.Before(todayStart) {
		tomorrowStart := todayStart.AddDate(0, 0, 1)
		remaining := tomorrowStart.Sub(now)
		c.JSON(http.StatusConflict, gin.H{
			"error":         "daily reward already claimed",
			"next_claim_in": remaining.String(),
			"next_claim_at": tomorrowStart.Format(time.RFC3339),
		})
		return
	}

	if err := h.awardTickets(uid, reward, "daily", "Daily login reward"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to claim reward"})
		return
	}

	h.DB.Model(&user).Update("last_daily_claim", now)
	h.DB.First(&user, uid)

	c.JSON(http.StatusOK, gin.H{
		"message":  "daily reward claimed",
		"tickets":  user.Tickets,
		"rewarded": reward,
	})
}

func (h *RewardHandler) DistributeMonthly(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin only"})
		return
	}

	now := makassarNow()
	period := c.DefaultQuery("period", now.Format("2006-01"))
	limitStr := c.DefaultQuery("limit", "10")
	var limit int
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit < 1 {
		limit = 10
	}

	reward := h.Config.Get("monthly_leaderboard")

	var topUsers []struct {
		ID uint
		XP int64
	}
	if err := h.DB.Raw(`
		SELECT id, xp FROM users
		WHERE xp > 0
		ORDER BY xp DESC
		LIMIT ?
	`, limit).Scan(&topUsers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	var results []gin.H
	for _, u := range topUsers {
		note := "Monthly XP leaderboard reward (" + period + ")"
		if err := h.awardTickets(u.ID, reward, "monthly_xp", note); err != nil {
			continue
		}
		results = append(results, gin.H{
			"user_id": u.ID,
			"tickets": reward,
		})
	}

	h.DB.Exec(`UPDATE users SET xp = 0 WHERE xp > 0`)

	c.JSON(http.StatusOK, gin.H{
		"message":        "monthly rewards distributed",
		"period":         period,
		"users_rewarded": len(results),
		"results":        results,
	})
}

func (h *RewardHandler) Status(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user model.User
	if err := h.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	todayStart := todayMakassarBoundary()

	canClaim := user.LastDailyClaim == nil || user.LastDailyClaim.Before(todayStart)
	var nextClaimAt string
	if !canClaim {
		tomorrowStart := todayStart.AddDate(0, 0, 1)
		nextClaimAt = tomorrowStart.Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, gin.H{
		"daily_reward": gin.H{
			"can_claim":    canClaim,
			"reward":       h.Config.Get("daily_reward"),
			"next_claim_at": nextClaimAt,
		},
	})
}
