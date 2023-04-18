package models

type Need struct {
	ID          uint   `gorm:"column:id"`
	Title       string `gorm:"column:title"`
	Amount      int    `gorm:"column:amount"`
	HelpEventID uint   `gorm:"column:help_event_id"`
	Unit        Unit   `gorm:"column:unit"`
}
