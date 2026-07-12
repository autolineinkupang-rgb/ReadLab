package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"wtr-lab-clone/backend/internal/config"
	"wtr-lab-clone/backend/internal/model"

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

var sampleNovels = []struct {
	Title       string
	AltTitle    string
	Slug        string
	Author      string
	AuthorSlug  string
	Status      string
	Description string
	Genres      []string
	Chapters    int
	Chars       string
	AIPercent   string
	Rating      float64
}{
	{
		Title: "Having Dinner with His Brother, the Cold and Aloof Tycoon Becomes Addicted to His Doting Affections",
		AltTitle: "陪哥哥吃饭，冷欲大佬强宠上瘾",
		Slug: "having-dinner-with-his-brother-the-cold-and-aloof-tycoon-becomes-addicted-to-his-doting-affections",
		Author: "半条活鱼", AuthorSlug: "ban-tiao-huo-yu",
		Status: "completed",
		Description: "[Cold and aloof tycoon × Bright and delicate, lazy princess + Forced marriage + 12-year age gap + 1v1, both are virgins]",
		Genres: []string{"romance", "slice-of-life", "urban-life"},
		Chapters: 135, Chars: "250K", AIPercent: "37%", Rating: 3.5,
	},
	{
		Title: "Corpse Puppet Phoenix Girl",
		AltTitle: "尸傀凰女",
		Slug: "corpse-puppet-phoenix-girl",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "ongoing",
		Description: "Meeting you at the most beautiful street corner was the worst decision I ever made.",
		Genres: []string{"adult", "adventure", "fantasy", "romance"},
		Chapters: 242, Chars: "653K", AIPercent: "20.7%", Rating: 3.8,
	},
	{
		Title: "Reborn As the Little Delicate Wife of the Domineering Ceo",
		AltTitle: "豪门重生，夫人超超超厉害",
		Slug: "reborn-as-the-little-delicate-wife-of-the-domineering-ceo",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "completed",
		Description: "Sweet and fluffy, incredibly romantic, the female lead's various powerful personas are exposed online.",
		Genres: []string{"romance", "urban-life"},
		Chapters: 378, Chars: "638K", AIPercent: "13.2%", Rating: 4.0,
	},
	{
		Title: "The Corpse Family is Heavy",
		AltTitle: "尸家重地",
		Slug: "the-corpse-family-is-heavy",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "completed",
		Description: "You can't be greedy for cheap deals in this world.",
		Genres: []string{"action", "adult", "fantasy", "mystery", "supernatural"},
		Chapters: 252, Chars: "469K", AIPercent: "19.8%", Rating: 3.2,
	},
	{
		Title: "Can You Please Comfort Me?",
		AltTitle: "可不可以哄哄我",
		Slug: "can-you-please-comfort-me",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "completed",
		Description: "As a child, Shen Shengsheng played house and made a promise with an older boy.",
		Genres: []string{"drama", "josei", "romance", "tragedy", "urban-life"},
		Chapters: 149, Chars: "235K", AIPercent: "33.6%", Rating: 3.0,
	},
	{
		Title: "First-rank Di Consort",
		AltTitle: "一品嫡妃",
		Slug: "first-rank-di-consort",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "completed",
		Description: "Song Anran, a wealthy and beautiful woman, is ambitiously expanding her business empire.",
		Genres: []string{"historical", "romance"},
		Chapters: 387, Chars: "2.99M", AIPercent: "15.5%", Rating: 3.6,
	},
	{
		Title: "I am the Crown Prince of the Ming Dynasty",
		AltTitle: "我在大明当太子",
		Slug: "i-am-the-crown-prince-of-the-ming-dynasty",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "completed",
		Description: "College student Zhu Yu transmigrates into the body of the famous Third Prince Zhu.",
		Genres: []string{"action", "adult", "adventure", "drama", "fantasy", "historical", "military", "xuanhuan"},
		Chapters: 1592, Chars: "2.96M", AIPercent: "3.14%", Rating: 4.2,
	},
	{
		Title: "Could I Really End Up 'collapsing My Image' Even in the World of Rule Horror",
		AltTitle: "我还能在规则怪谈里塌房不成？",
		Slug: "could-i-really-end-up-collapsing-my-image-even-in-the-world-of-rule-horror",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "ongoing",
		Description: "Infinite Flow + Rule-Based Ghost Stories + Thriller Game + Strong Female Team.",
		Genres: []string{"mystery", "psychological"},
		Chapters: 925, Chars: "1.75M", AIPercent: "5.4%", Rating: 4.1,
	},
	{
		Title: "The Legend of the Mountain and Sea Demon Subduing",
		AltTitle: "大丰小道士",
		Slug: "the-legend-of-the-mountain-and-sea-demon-subduing",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "completed",
		Description: "In the realm of mountains and seas, sects stand in great numbers, and demons roam freely.",
		Genres: []string{"action", "adventure", "martial-arts"},
		Chapters: 1522, Chars: "2.82M", AIPercent: "3.29%", Rating: 3.9,
	},
	{
		Title: "Don't Be Too Wild",
		AltTitle: "别太野",
		Slug: "dont-be-too-wild",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "ongoing",
		Description: "A seemingly innocent but actually rebellious heiress x a cold and roguish prince.",
		Genres: []string{"romance", "school-life", "slice-of-life"},
		Chapters: 160, Chars: "309K", AIPercent: "31.3%", Rating: 3.4,
	},
	{
		Title: "Naruto: In Konoha Village, I Awakened Wood Release at the Start",
		AltTitle: "火影：木叶村，开局觉醒木遁",
		Slug: "naruto-in-konoha-village-i-awakened-wood-release-at-the-start",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "ongoing",
		Description: "Konoha 52nd year. Chiba awakened her memories of her past life.",
		Genres: []string{"action", "fan-fiction", "fantasy", "martial-arts", "seinen", "supernatural"},
		Chapters: 1002, Chars: "1.8M", AIPercent: "8.5%", Rating: 3.8,
	},
	{
		Title: "I Just Started High School, But the System Insists I'm an Emperor in My Twilight Years",
		AltTitle: "刚上高一，系统非说我是晚年大帝",
		Slug: "i-just-started-high-school-but-the-system-insists-im-an-emperor-in-my-twilight-years",
		Author: "佚名", AuthorSlug: "yi-ming",
		Status: "ongoing",
		Description: "Jiang Feng is an ordinary high school student at Linchuan High School.",
		Genres: []string{"action", "comedy", "fantasy", "martial-arts", "school-life", "supernatural", "urban-life"},
		Chapters: 264, Chars: "450K", AIPercent: "12%", Rating: 1.9,
	},
	{
		Title: "The Wizard's Secret Library",
		AltTitle: "巫师的神秘图书馆",
		Slug: "the-wizards-secret-library",
		Author: "墨色", AuthorSlug: "mo-se",
		Status: "ongoing",
		Description: "A librarian discovers a hidden section in the ancient library that contains real magic books. As he learns wizardry, he gets entangled in a hidden magical world.",
		Genres: []string{"action", "adventure", "fantasy", "mystery", "romance"},
		Chapters: 687, Chars: "1.2M", AIPercent: "5.5%", Rating: 4.3,
	},
	{
		Title: "Rebirth of the Heavenly Chef",
		AltTitle: "重生之神厨",
		Slug: "rebirth-of-the-heavenly-chef",
		Author: "酸甜排骨", AuthorSlug: "suan-tian-pai-gu",
		Status: "completed",
		Description: "A Michelin-star chef is reborn into a poor family in ancient times. Using his modern culinary knowledge, he rises from a street stall to become the imperial chef.",
		Genres: []string{"comedy", "drama", "historical", "slice-of-life"},
		Chapters: 856, Chars: "2.1M", AIPercent: "4.8%", Rating: 4.1,
	},
	{
		Title: "My Girlfriend is a Virtual Idol",
		AltTitle: "我的女友是虚拟偶像",
		Slug: "my-girlfriend-is-a-virtual-idol",
		Author: "星河", AuthorSlug: "xing-he",
		Status: "ongoing",
		Description: "A programmer accidentally creates an AI that becomes a popular virtual idol. Only he knows the truth behind the screen. A sweet sci-fi romance.",
		Genres: []string{"comedy", "romance", "school-life", "sci-fi", "slice-of-life"},
		Chapters: 423, Chars: "890K", AIPercent: "2.3%", Rating: 4.0,
	},
	{
		Title: "The Last Necromancer",
		AltTitle: "最后的死灵法师",
		Slug: "the-last-necromancer",
		Author: "幽冥", AuthorSlug: "you-ming",
		Status: "completed",
		Description: "In a world that despises necromancy, the last practitioner must hide his powers while uncovering a conspiracy that threatens both the living and the dead.",
		Genres: []string{"action", "adventure", "drama", "fantasy", "mystery", "supernatural"},
		Chapters: 1245, Chars: "2.8M", AIPercent: "7.2%", Rating: 4.5,
	},
	{
		Title: "Transmigrated as the Villain's Mother",
		AltTitle: "穿成反派的娘亲",
		Slug: "transmigrated-as-the-villains-mother",
		Author: "轻舞", AuthorSlug: "qing-wu",
		Status: "ongoing",
		Description: "She transmigrates into a novel as the mother of the future villain. Now she must raise the child properly to avoid the tragic ending.",
		Genres: []string{"comedy", "drama", "fantasy", "josei", "romance", "slice-of-life"},
		Chapters: 567, Chars: "1.1M", AIPercent: "9.1%", Rating: 3.9,
	},
	{
		Title: "Dungeon Architect",
		AltTitle: "地下城建筑师",
		Slug: "dungeon-architect",
		Author: "石中剑", AuthorSlug: "shi-zhong-jian",
		Status: "ongoing",
		Description: "Zhao Yun is transported to a fantasy world where he must design and manage dungeons for adventurers. But his modern engineering knowledge turns simple dungeons into death traps.",
		Genres: []string{"action", "adventure", "comedy", "fantasy", "game"},
		Chapters: 789, Chars: "1.5M", AIPercent: "3.8%", Rating: 4.2,
	},
	{
		Title: "The Heiress's Secret Bodyguard",
		AltTitle: "千金大小姐的秘密保镖",
		Slug: "the-heiresss-secret-bodyguard",
		Author: "夜雨", AuthorSlug: "ye-yu",
		Status: "completed",
		Description: "A retired special forces soldier takes a job as a bodyguard for a wealthy heiress. But her life is more dangerous than any battlefield he's faced.",
		Genres: []string{"action", "adult", "drama", "romance", "urban-life"},
		Chapters: 456, Chars: "890K", AIPercent: "11.5%", Rating: 3.7,
	},
	{
		Title: "I Can See Game Panels",
		AltTitle: "我能看见游戏面板",
		Slug: "i-can-see-game-panels",
		Author: "数据帝", AuthorSlug: "shu-ju-di",
		Status: "ongoing",
		Description: "Chen Ming wakes up one day able to see status panels above everyone's heads. Now he can optimize his life like an RPG character.",
		Genres: []string{"comedy", "fantasy", "game", "school-life", "urban-life"},
		Chapters: 334, Chars: "670K", AIPercent: "6.7%", Rating: 3.6,
	},
	{
		Title: "Martial Arts Online",
		AltTitle: "武道 online",
		Slug: "martial-arts-online",
		Author: "剑客", AuthorSlug: "jian-ke",
		Status: "ongoing",
		Description: "A full-dive VR game based on Chinese martial arts attracts millions of players. But when players start getting hurt in real life, the game becomes a matter of survival.",
		Genres: []string{"action", "adventure", "fantasy", "game", "martial-arts", "mystery", "sci-fi"},
		Chapters: 1102, Chars: "2.3M", AIPercent: "4.2%", Rating: 4.4,
	},
	{
		Title: "The Demon Lord is a Pop Star",
		AltTitle: "魔王是流行明星",
		Slug: "the-demon-lord-is-a-pop-star",
		Author: "霓虹", AuthorSlug: "ni-hong",
		Status: "ongoing",
		Description: "The Demon Lord is banished to Earth and must find a way to survive. He accidentally becomes a K-pop idol. Now he must balance world domination with dance practice.",
		Genres: []string{"comedy", "fantasy", "music", "romance", "school-life", "slice-of-life"},
		Chapters: 298, Chars: "540K", AIPercent: "15.3%", Rating: 4.1,
	},
	{
		Title: "Apocalypse: I Have a Farm",
		AltTitle: "末世：我有一个农场",
		Slug: "apocalypse-i-have-a-farm",
		Author: "种田人", AuthorSlug: "zhong-tian-ren",
		Status: "ongoing",
		Description: "When the apocalypse strikes, everyone is fighting for resources. But Li Wei has a magical farm that produces unlimited supplies. Building a sanctuary in the wasteland.",
		Genres: []string{"action", "adventure", "drama", "fantasy", "sci-fi", "slice-of-life"},
		Chapters: 967, Chars: "1.9M", AIPercent: "3.5%", Rating: 4.0,
	},
	{
		Title: "The Blind Swordsman",
		AltTitle: "盲剑客",
		Slug: "the-blind-swordsman",
		Author: "独孤", AuthorSlug: "du-gu",
		Status: "completed",
		Description: "A blind martial artist roams the jianghu, relying on his heightened senses to detect danger. His swordsmanship is legendary, but his past is shrouded in mystery.",
		Genres: []string{"action", "adventure", "drama", "historical", "martial-arts", "seinen", "tragedy"},
		Chapters: 678, Chars: "1.3M", AIPercent: "1.2%", Rating: 4.6,
	},
	{
		Title: "My Neighbor is a Time Traveler",
		AltTitle: "我的邻居是穿越者",
		Slug: "my-neighbor-is-a-time-traveler",
		Author: "时光", AuthorSlug: "shi-guang",
		Status: "completed",
		Description: "A man discovers his new neighbor is a time traveler from the future. Together they navigate paradoxes, prevent disasters, and deal with annoying landlord issues.",
		Genres: []string{"comedy", "drama", "romance", "sci-fi", "slice-of-life", "urban-life"},
		Chapters: 234, Chars: "420K", AIPercent: "8.9%", Rating: 3.8,
	},
	{
		Title: "Underground Fighting King",
		AltTitle: "地下拳王",
		Slug: "underground-fighting-king",
		Author: "铁拳", AuthorSlug: "tie-quan",
		Status: "ongoing",
		Description: "A young man enters the brutal world of underground fighting to pay off his family's debts. With each victory, he climbs higher but the danger grows.",
		Genres: []string{"action", "adult", "drama", "martial-arts", "psychological", "urban-life"},
		Chapters: 534, Chars: "1.0M", AIPercent: "10.4%", Rating: 3.5,
	},
	{
		Title: "The Ghost King's Beloved",
		AltTitle: "鬼王的心尖宠",
		Slug: "the-ghost-kings-beloved",
		Author: "灵异", AuthorSlug: "ling-yi",
		Status: "completed",
		Description: "A young exorcist accidentally binds herself to the Ghost King. Now she must navigate the spirit world while dealing with a possessive and powerful ghost husband.",
		Genres: []string{"adventure", "comedy", "fantasy", "romance", "supernatural", "xianxia"},
		Chapters: 445, Chars: "780K", AIPercent: "18.7%", Rating: 3.4,
	},
	{
		Title: "Infinite Loop Gamer",
		AltTitle: "无限循环玩家",
		Slug: "infinite-loop-gamer",
		Author: "轮回", AuthorSlug: "lun-hui",
		Status: "ongoing",
		Description: "Su Ming is trapped in an infinite loop of horror games. Each death resets the loop, but the horrors evolve. He must use every death as a lesson to finally escape.",
		Genres: []string{"action", "adventure", "horror", "mystery", "psychological", "sci-fi", "supernatural"},
		Chapters: 890, Chars: "1.7M", AIPercent: "6.1%", Rating: 4.3,
	},
	{
		Title: "CEO's Substitute Bride",
		AltTitle: "总裁的替身新娘",
		Slug: "ceos-substitute-bride",
		Author: "蔷薇", AuthorSlug: "qiang-wei",
		Status: "completed",
		Description: "A poor girl is forced to marry a cold CEO as a substitute for her runaway sister. She expects a loveless marriage, but slowly melts his frozen heart.",
		Genres: []string{"drama", "josei", "romance", "urban-life"},
		Chapters: 345, Chars: "620K", AIPercent: "22.5%", Rating: 3.3,
	},
	{
		Title: "The Alchemist's Apprentice",
		AltTitle: "炼金术士的学徒",
		Slug: "the-alchemists-apprentice",
		Author: "黄金", AuthorSlug: "huang-jin",
		Status: "ongoing",
		Description: "A poor orphan becomes the apprentice of a mysterious alchemist. As he learns the secrets of transmutation, he discovers that alchemy is more science than magic.",
		Genres: []string{"adventure", "drama", "fantasy", "mystery", "sci-fi"},
		Chapters: 612, Chars: "1.1M", AIPercent: "4.5%", Rating: 4.0,
	},
	{
		Title: "Zombie Apocalypse: Building a Base",
		AltTitle: "丧尸危机：建造基地",
		Slug: "zombie-apocalypse-building-a-base",
		Author: "幸存者", AuthorSlug: "xing-cun-zhe",
		Status: "ongoing",
		Description: "When the zombie apocalypse breaks out, an architect uses his knowledge to build an impenetrable base. But surviving humans are often more dangerous than zombies.",
		Genres: []string{"action", "adult", "drama", "horror", "psychological", "sci-fi"},
		Chapters: 756, Chars: "1.4M", AIPercent: "8.3%", Rating: 3.9,
	},
	{
		Title: "The Fox Spirit's Promise",
		AltTitle: "狐妖的承诺",
		Slug: "the-fox-spirits-promise",
		Author: "九尾", AuthorSlug: "jiu-wei",
		Status: "completed",
		Description: "A thousand-year-old fox spirit makes a promise to protect a reincarnated lover. But in modern times, he's a struggling artist who doesn't believe in the supernatural.",
		Genres: []string{"drama", "fantasy", "historical", "romance", "supernatural", "tragedy", "xianxia"},
		Chapters: 523, Chars: "950K", AIPercent: "12.8%", Rating: 4.2,
	},
	{
		Title: "System Overlord: Mecha Wars",
		AltTitle: "系统霸主：机甲大战",
		Slug: "system-overlord-mecha-wars",
		Author: "机师", AuthorSlug: "ji-shi",
		Status: "ongoing",
		Description: "In the year 3050, mecha pilots fight for control of the galaxy. A rookie pilot awakens a system that lets him upgrade his mecha beyond known limits.",
		Genres: []string{"action", "adventure", "game", "mecha", "military", "sci-fi"},
		Chapters: 1100, Chars: "2.5M", AIPercent: "2.9%", Rating: 4.4,
	},
}

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Genre{},
		&model.Tag{},
		&model.Novel{},
		&model.NovelGenre{},
		&model.Chapter{},
		&model.User{},
		&model.Vote{},
		&model.Request{},
		&model.TicketTransaction{},
		&model.News{},
		&model.ReadingHistory{},
		&model.NovelFollow{},
		&model.Share{},
	); err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}

	seedGenres(db)
	seedTags(db)
	seedUsers(db)
	seedNews(db)
	seedRatings(db)

	log.Println("seed completed")
}

func seedGenres(db *gorm.DB) {
	for _, g := range genres {
		db.FirstOrCreate(&model.Genre{}, model.Genre{Slug: g.Slug, Name: g.Name})
	}
	fmt.Printf("seeded %d genres\n", len(genres))
}

func seedTags(db *gorm.DB) {
	for _, t := range tags {
		db.FirstOrCreate(&model.Tag{}, model.Tag{Slug: t.Slug, Name: t.Name})
	}
	fmt.Printf("seeded %d tags\n", len(tags))
}

func seedNovels(db *gorm.DB) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i, s := range sampleNovels {
		novel := model.Novel{
			Title:       s.Title,
			AltTitle:    s.AltTitle,
			Slug:        s.Slug,
			Author:      s.Author,
			AuthorSlug:  s.AuthorSlug,
			Status:      s.Status,
			Views:       uint64(rng.Intn(500000) + 1000),
			Rating:      s.Rating,
			RatingCount: uint(rng.Intn(500) + 20),
			Chapters:    s.Chapters,
			Readers:     rng.Intn(5000) + 10,
			Chars:       s.Chars,
			AIPercent:   s.AIPercent,
			Description: s.Description,
		}

		var existing model.Novel
		if err := db.Where("slug = ?", novel.Slug).First(&existing).Error; err == nil {
			continue
		}

		if err := db.Create(&novel).Error; err != nil {
			log.Printf("failed to create novel %s: %v", novel.Title, err)
			continue
		}

		for _, genreSlug := range s.Genres {
			var genre model.Genre
			if err := db.Where("slug = ?", genreSlug).First(&genre).Error; err == nil {
				db.Create(&model.NovelGenre{NovelID: novel.ID, GenreID: genre.ID})
			}
		}

		seedChapters(db, novel.ID, s.Chapters)

		fmt.Printf("seeded novel %d/%d: %s (%d chapters)\n", i+1, len(sampleNovels), novel.Title, s.Chapters)
	}
}

func seedChapters(db *gorm.DB, novelID uint, count int) {
	for i := 1; i <= count; i++ {
		chapter := model.Chapter{
			NovelID:  novelID,
			Number:   i,
			Title:    fmt.Sprintf("Chapter %d", i),
			Content:  fmt.Sprintf("This is the content of chapter %d. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", i),
			IsLocked: i > count/2,
		}

		var existing model.Chapter
		if err := db.Where("novel_id = ? AND number = ?", novelID, i).First(&existing).Error; err == nil {
			continue
		}

		db.Create(&chapter)
	}
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
