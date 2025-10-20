package models

// explicit join model to match existing migration (has its own id)
type MangaGenre struct {
    ID      int64 `json:"id" gorm:"primaryKey;autoIncrement"`
    MangaID int64 `json:"manga_id" gorm:"index;not null"`
    GenreID int64 `json:"genre_id" gorm:"index;not null"`
}

func (MangaGenre) TableName() string {
    return "manga_genres"
}