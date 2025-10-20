package models

type Genre struct {
    ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
    Name string `json:"name" gorm:"unique;not null"`
}

func (Genre) TableName() string {
    return "genres"
}