package service

import (
	"math"

	"gorm.io/gorm"
	"wtr-lab-clone/backend/internal/model"
)

type NovelService struct {
	DB *gorm.DB
}

func NewNovelService(db *gorm.DB) *NovelService {
	return &NovelService{DB: db}
}

type NovelFilter struct {
	Page   int
	Limit  int
	Sort   string
	Order  string
	Status string
	Genre  string
	Search string
}

type NovelPage struct {
	Data       []model.Novel
	Page       int
	Limit      int
	Total      int64
	TotalPages int
}

func (s *NovelService) List(filter NovelFilter) (*NovelPage, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	validSorts := map[string]bool{
		"created_at": true, "title": true, "views": true,
		"readers": true, "chapters": true, "rating": true, "votes": true,
	}
	if !validSorts[filter.Sort] {
		filter.Sort = "created_at"
	}
	if filter.Order != "asc" && filter.Order != "desc" {
		filter.Order = "desc"
	}

	query := s.DB.Model(&model.Novel{}).Preload("Genres")

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Genre != "" {
		sub := s.DB.Table("novel_genres").
			Select("novel_id").
			Where("genre_id = (SELECT id FROM genres WHERE slug = ?)", filter.Genre)
		query = query.Where("id IN (?)", sub)
	}
	if filter.Search != "" {
		query = query.Where("title ILIKE ? OR author ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	var total int64
	query.Count(&total)

	offset := (filter.Page - 1) * filter.Limit
	orderClause := filter.Sort + " " + filter.Order

	var novels []model.Novel
	if err := query.Order(orderClause).Offset(offset).Limit(filter.Limit).Find(&novels).Error; err != nil {
		return nil, err
	}

	return &NovelPage{
		Data:       novels,
		Page:       filter.Page,
		Limit:      filter.Limit,
		Total:      total,
		TotalPages: int(math.Ceil(float64(total) / float64(filter.Limit))),
	}, nil
}

func (s *NovelService) Trending(limit int) ([]model.Novel, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	var novels []model.Novel
	err := s.DB.Preload("Genres").Order("views DESC").Limit(limit).Find(&novels).Error
	return novels, err
}

func (s *NovelService) Recommendations(limit int) ([]model.Novel, error) {
	if limit < 1 || limit > 100 {
		limit = 12
	}
	var novels []model.Novel
	err := s.DB.Preload("Genres").Where("rating > 0").Order("rating DESC").Limit(limit).Find(&novels).Error
	return novels, err
}

func (s *NovelService) Random(limit int) ([]model.Novel, error) {
	if limit < 1 || limit > 100 {
		limit = 1
	}
	var maxID int64
	s.DB.Model(&model.Novel{}).Select("MAX(id)").Scan(&maxID)
	if maxID == 0 {
		return nil, nil
	}

	var novels []model.Novel
	err := s.DB.Preload("Genres").
		Where("id >= FLOOR(RANDOM() * ?) + 1", maxID).
		Limit(limit).
		Find(&novels).Error
	return novels, err
}

func (s *NovelService) GetByID(id uint) (*model.Novel, error) {
	var novel model.Novel
	err := s.DB.Preload("Genres").First(&novel, id).Error
	if err != nil {
		return nil, err
	}
	return &novel, nil
}

func (s *NovelService) GetBySlug(slug string) (*model.Novel, error) {
	var novel model.Novel
	err := s.DB.Where("slug = ?", slug).Preload("Genres").First(&novel).Error
	if err != nil {
		return nil, err
	}
	return &novel, nil
}

func (s *NovelService) GetChapters(novelID uint) ([]model.Chapter, error) {
	var chapters []model.Chapter
	err := s.DB.Where("novel_id = ?", novelID).Order("number ASC").Find(&chapters).Error
	return chapters, err
}

func (s *NovelService) GetChapterByNumber(novelID uint, number int) (*model.Chapter, error) {
	var chapter model.Chapter
	err := s.DB.Where("novel_id = ? AND number = ?", novelID, number).First(&chapter).Error
	if err != nil {
		return nil, err
	}
	return &chapter, nil
}

func (s *NovelService) Search(q string, page, limit int) ([]model.Novel, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	var total int64
	s.DB.Model(&model.Novel{}).
		Where("title ILIKE ? OR author ILIKE ?", "%"+q+"%", "%"+q+"%").
		Count(&total)

	var novels []model.Novel
	offset := (page - 1) * limit
	err := s.DB.Preload("Genres").
		Where("title ILIKE ? OR author ILIKE ?", "%"+q+"%", "%"+q+"%").
		Order("views DESC").
		Offset(offset).Limit(limit).
		Find(&novels).Error

	return novels, total, err
}

type AutocompleteResult struct {
	ID    uint   `json:"id"`
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

func (s *NovelService) Autocomplete(q string, limit int) ([]AutocompleteResult, error) {
	if limit < 1 || limit > 20 {
		limit = 5
	}
	var results []AutocompleteResult
	err := s.DB.Model(&model.Novel{}).
		Select("id, slug, title").
		Where("title ILIKE ?", "%"+q+"%").
		Order("views DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}
