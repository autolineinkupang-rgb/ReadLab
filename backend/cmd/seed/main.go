package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"readlab/backend/internal/config"
	"readlab/backend/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var tags = []struct {
	Slug string
	Name string
}{
	{"male-protagonist", "Male Protagonist"},
	{"female-protagonist", "Female Protagonist"},
	{"transmigration", "Transmigration"},
	{"reincarnation", "Reincarnation"},
	{"system", "System"},
	{"cultivation", "Cultivation"},
	{"overpowered-protagonist", "Overpowered Protagonist"},
	{"weak-to-strong", "Weak to Strong"},
	{"harem", "Harem"},
	{"love-triangle", "Love Triangle"},
	{"ancient-world", "Ancient World"},
	{"modern-world", "Modern World"},
	{"game-world", "Game World"},
	{"adaptation", "Adaptation"},
	{"slow-burn", "Slow Burn"},
	{"revenge", "Revenge"},
	{"superpowers", "Superpowers"},
	{"magic-school", "Magic School"},
	{"military", "Military"},
	{"smart-protagonist", "Smart Protagonist"},
	{"antihero-protagonist", "Antihero Protagonist"},
	{"apocalypse", "Apocalypse"},
	{"survival", "Survival"},
	{"post-apocalyptic", "Post-Apocalyptic"},
	{"conspiracy", "Conspiracy"},
	{"politics", "Politics"},
	{"business", "Business"},
	{"showbiz", "Showbiz"},
	{"time-travel", "Time Travel"},
	{"multiple-worlds", "Multiple Worlds"},
	{"quick-transmigration", "Quick Transmigration"},
	{"villain-protagonist", "Villain Protagonist"},
	{"comedy", "Comedy"},
	{"tragedy", "Tragedy"},
	{"slice-of-life", "Slice of Life"},
	{"school-life", "School Life"},
	{"office-life", "Office Life"},
	{"royalty", "Royalty"},
	{"noblesse-oblige", "Noblesse Oblige"},
	{"master-servant", "Master Servant"},
	{"yandere", "Yandere"},
	{"tsundere", "Tsundere"},
	{"amnesia", "Amnesia"},
	{"identity-hidden", "Identity Hidden"},
	{"secret-identity", "Secret Identity"},
	{"body-swap", "Body Swap"},
	{"gender-bender", "Gender Bender"},
	{"cross-dressing", "Cross Dressing"},
	{"monster-protagonist", "Monster Protagonist"},
	{"ghost-protagonist", "Ghost Protagonist"},
}

var genres = []struct {
	Slug string
	Name string
}{
	{"action", "Action"},
	{"adult", "Adult"},
	{"adventure", "Adventure"},
	{"comedy", "Comedy"},
	{"drama", "Drama"},
	{"ecchi", "Ecchi"},
	{"erciyuan", "Erciyuan"},
	{"fan-fiction", "Fan-Fiction"},
	{"fantasy", "Fantasy"},
	{"game", "Game"},
	{"gender-bender", "Gender Bender"},
	{"harem", "Harem"},
	{"historical", "Historical"},
	{"horror", "Horror"},
	{"josei", "Josei"},
	{"martial-arts", "Martial Arts"},
	{"mature", "Mature"},
	{"mecha", "Mecha"},
	{"military", "Military"},
	{"mystery", "Mystery"},
	{"psychological", "Psychological"},
	{"romance", "Romance"},
	{"school-life", "School Life"},
	{"sci-fi", "Sci-Fi"},
	{"seinen", "Seinen"},
	{"shoujo", "Shoujo"},
	{"shoujo-ai", "Shoujo Ai"},
	{"shounen", "Shounen"},
	{"shounen-ai", "Shounen Ai"},
	{"slice-of-life", "Slice of Life"},
	{"smut", "Smut"},
	{"sports", "Sports"},
	{"supernatural", "Supernatural"},
	{"tragedy", "Tragedy"},
	{"urban-life", "Urban Life"},
	{"wuxia", "Wuxia"},
	{"xianxia", "Xianxia"},
	{"xuanhuan", "Xuanhuan"},
	{"yaoi", "Yaoi"},
	{"yuri", "Yuri"},
}

func seedUsers(db *gorm.DB) {
	users := []struct {
		Username string
		Email    string
		Password string
		Tickets  float64
		IsAdmin  bool
	}{
		{"Mega_bells", "mega@example.com", "password", 3569.76, false},
		{"StandardCrystal", "crystal@example.com", "password", 2907.17, false},
		{"Alpha2", "alpha2@example.com", "password", 2693.07, false},
		{"reader1", "reader1@example.com", "password", 100, false},
	}

	for _, u := range users {
		var existing model.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			continue
		}

		hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		role := "member"
		if u.IsAdmin {
			role = "admin"
		}
		user := model.User{
			Username:     u.Username,
			Email:        u.Email,
			PasswordHash: string(hash),
			DisplayName:  u.Username,
			Tickets:      u.Tickets,
			Role:         role,
		}
		db.Create(&user)
		fmt.Printf("seeded user: %s\n", u.Username)
	}
}

func seedRatings(db *gorm.DB) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var novels []model.Novel
	db.Where("rating = 0 OR rating_count = 0").Find(&novels)

	for _, novel := range novels {
		// realistic rating distribution: most between 3.5-4.5, some higher/lower
		rating := 3.0 + rng.Float64()*2.0
		if rating > 4.8 {
			rating = 4.8
		}
		rating = float64(int(rating*10)) / 10 // round to 1 decimal

		ratingCount := uint(rng.Intn(2000) + 10)
		if novel.Views == 0 {
			novel.Views = uint64(rng.Intn(500000) + 1000)
		}
		if novel.Readers == 0 {
			novel.Readers = rng.Intn(5000) + 10
		}
		if novel.Votes == 0 {
			novel.Votes = uint(rng.Intn(500) + 1)
		}

		db.Model(&novel).Updates(map[string]interface{}{
			"rating":       rating,
			"rating_count": ratingCount,
			"views":        novel.Views,
			"readers":      novel.Readers,
			"votes":        novel.Votes,
		})
	}

	fmt.Printf("seeded ratings for %d novels\n", len(novels))
}

func seedNews(db *gorm.DB) {
	news := []struct {
		Title   string
		Content string
		Type    string
		Slug    string
	}{
		{
			Title: "🎉 16th Giveaway Winners 🎉",
			Content: "Congratulations to all the winners of our 16th Giveaway!",
			Type: "news", Slug: "16th-giveaway-winners",
		},
		{
			Title: "🎉 Our 16th Giveaway is LIVE! 🎉",
			Content: "Our 16th Giveaway is now live! Participate now for a chance to win amazing prizes.",
			Type: "news", Slug: "16th-giveaway-live",
		},
		{
			Title: "Version 1.13.3 - New Source Management & Bug Fixes!",
			Content: "We have released version 1.13.3 with new source management features and various bug fixes.",
			Type: "changelog", Slug: "v1-13-3",
		},
	}

	for _, n := range news {
		var existing model.News
		if err := db.Where("slug = ?", n.Slug).First(&existing).Error; err == nil {
			continue
		}

		db.Create(&model.News{
			Title:   n.Title,
			Content: n.Content,
			Type:    n.Type,
			Slug:    n.Slug,
		})
		fmt.Printf("seeded news: %s\n", n.Title)
	}
}
