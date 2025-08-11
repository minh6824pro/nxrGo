package dto

type CategoryResponse struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"type:varchar(255);unique" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
}
