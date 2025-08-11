package dto

type MerchantResponse struct {
	ID   uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name string `gorm:"type:varchar(255);unique" json:"name" binding:"required"  `
}
