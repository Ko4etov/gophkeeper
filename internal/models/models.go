package models

import (
	"time"
)

// DataType определяет тип хранимых данных
type DataType string

const (
	TypeLoginPassword DataType = "login_password"
	TypeText          DataType = "text"
	TypeBinary        DataType = "binary"
	TypeBankCard      DataType = "bank_card"
)

// DataMeta содержит метаданные для любого типа данных
type DataMeta struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   time.Time `json:"deleted_at,omitempty"`
	SyncedAt    time.Time `json:"synced_at,omitempty"`
	Version     int       `json:"version"`
	DataType    DataType  `json:"data_type"`
}

// DataEntry представляет полную запись данных
type DataEntry struct {
	Meta DataMeta    `json:"meta"`
	Data interface{} `json:"data"`
}

// LoginPasswordData представляет данные типа "логин/пароль"
type LoginPasswordData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URL      string `json:"url,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

// TextData представляет произвольные текстовые данные
type TextData struct {
	Content string `json:"content"`
	Format  string `json:"format,omitempty"`
}

// BankCardData представляет данные банковской карты
type BankCardData struct {
	CardNumber  string `json:"card_number"`
	CardHolder  string `json:"card_holder"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
	CVV         string `json:"cvv,omitempty"`
	CardType    string `json:"card_type,omitempty"`
	BankName    string `json:"bank_name,omitempty"`
}

type BinaryData struct {
    Filename string `json:"filename"`
    Content  []byte `json:"content"`
    MimeType string `json:"mime_type,omitempty"`
    Size     int64  `json:"size"`
}
