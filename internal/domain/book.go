package domain

type Book struct {
	// ID        uint `gorm:"primarykey"`
	// CreatedAt time.Time
	// UpdatedAt time.Time
	ID    int    `json:"id"`
	Title string `json:"title"`
	// Authors []string `json:"authors"`
	Year int `json:"year"`
}
