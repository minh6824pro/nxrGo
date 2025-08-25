package dto

type BrandResponse struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);unique" json:"name"`
}
