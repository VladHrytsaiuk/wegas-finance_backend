package parsers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

func TestMonobankParser_ParseCSV(t *testing.T) {
	parser := NewMonobankParser()
	csvData := `Дата,Опис,MCC,Сума
01.01.2024 12:00:00,Сільпо,5411,-100.50
02.01.2024 15:30:00,Переказ від друга,,500.00`

	reader := strings.NewReader(csvData)
	transactions, err := parser.Parse(reader, int64(len(csvData)), "mono.csv")

	assert.NoError(t, err)
	assert.Len(t, transactions, 2)

	assert.Equal(t, int64(10050), transactions[0].Amount)
	assert.Equal(t, "expense", transactions[0].Type)
	assert.Equal(t, "5411", transactions[0].MCC)
	assert.Equal(t, "Сільпо", transactions[0].Description)

	assert.Equal(t, int64(50000), transactions[1].Amount)
	assert.Equal(t, "income", transactions[1].Type)
	assert.Equal(t, "", transactions[1].MCC)
}

func TestMonobankParser_ParseXLSX(t *testing.T) {
	parser := NewMonobankParser()
	
	f := excelize.NewFile()
	sheet := "Sheet1"
	f.SetCellValue(sheet, "A1", "Дата")
	f.SetCellValue(sheet, "B1", "Опис")
	f.SetCellValue(sheet, "C1", "MCC")
	f.SetCellValue(sheet, "D1", "Сума")
	
	f.SetCellValue(sheet, "A2", "01.01.2024 12:00:00")
	f.SetCellValue(sheet, "B2", "АТБ")
	f.SetCellValue(sheet, "C2", "5411")
	f.SetCellValue(sheet, "D2", "-50.25")

	var buf bytes.Buffer
	err := f.Write(&buf)
	assert.NoError(t, err)

	transactions, err := parser.Parse(&buf, int64(buf.Len()), "mono.xlsx")

	assert.NoError(t, err)
	assert.Len(t, transactions, 1)
	assert.Equal(t, int64(5025), transactions[0].Amount)
	assert.Equal(t, "5411", transactions[0].MCC)
	assert.Equal(t, "АТБ", transactions[0].Description)
}

func TestMonobankParser_UnsupportedFormat(t *testing.T) {
	parser := NewMonobankParser()
	reader := strings.NewReader("some data")
	_, err := parser.Parse(reader, 9, "unsupported.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "непідтримуваний формат файлу")
}
