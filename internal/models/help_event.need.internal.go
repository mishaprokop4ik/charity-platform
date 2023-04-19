package models

type Need struct {
	ID            uint   `gorm:"column:id"`
	Title         string `gorm:"column:title"`
	Amount        int    `gorm:"column:amount"`
	Received      int    `gorm:"column:received"`
	ReceivedTotal int    `gorm:"column:received_total"`
	HelpEventID   uint   `gorm:"column:help_event_id"`
	TransactionID *uint  `gorm:"column:transaction_id"`
	Unit          Unit   `gorm:"column:unit"`
}

func (Need) TableName() string {
	return "need"
}
