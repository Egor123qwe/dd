package model

import "time"

type Rent struct {
	ID            string     `json:"id"`
	UserID        string     `json:"-" db:"user_id"`
	Start         string     `json:"-"`
	End           string     `json:"-"`
	PriceByHour   float32    `json:"-" db:"price"`
	Cost          float32    `json:"cost" db:"cost"`
	Duration      int        `json:"duration"`       // минуты; при отрицательном в репо прибавляется 180
	Rating        *int       `json:"rating"`         // NULL в БД (нет отзыва) — сканируется как nil
	StartedAt     *time.Time `json:"started_at" db:"started_at"`     // дата/время начала сессии
	EndedAt       *time.Time `json:"ended_at" db:"ended_at"`         // дата/время окончания
	TemplateID    string     `json:"template_id" db:"template_id"`  // какой темплейт запускался
	TemplateTitle string     `json:"template_title" db:"template_title"` // название темплейта
}

type User struct {
	UserID string `json:"user_id"`
}

// AdminRent — аренда с указанием покупателя и поставщика для админ-панели.
type AdminRent struct {
	ID            string     `json:"id"`
	ClientID      string     `json:"client_id" db:"client_id"`
	MerchantID    string     `json:"merchant_id" db:"merchant_id"`
	Cost          float32    `json:"cost" db:"cost"`
	Duration      int        `json:"duration"`
	Rating        *int       `json:"rating"`
	StartedAt     *time.Time `json:"started_at" db:"started_at"`
	EndedAt       *time.Time `json:"ended_at" db:"ended_at"`
	TemplateID    string     `json:"template_id" db:"template_id"`
	TemplateTitle string     `json:"template_title" db:"template_title"`
}
