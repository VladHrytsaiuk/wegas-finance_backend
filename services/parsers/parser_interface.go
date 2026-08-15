package parsers

import (
	"io"
	"time"
)

type ParsedTransaction struct {
	Date             time.Time
	Amount           int64
	Description      string
	CounterpartyName string
	Type             string
	// TransferDirection is "out" when money leaves the statement account and
	// "in" when it returns from the savings account.
	TransferDirection string
	RawLine           string
	MCC               string // 🔥 Додаємо це поле
	BankCategory      string // 🔥 Додаємо поле для текстової категорії з виписки
}

type BankStatementParser interface {
	Parse(reader io.Reader, size int64, fileName string) ([]ParsedTransaction, error)
}
