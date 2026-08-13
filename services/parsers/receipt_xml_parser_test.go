package parsers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSilpoXMLParser_Parse(t *testing.T) {
	parser := NewSilpoXMLParser()

	raw := []byte(`<?xml version="1.0" encoding="windows-1251"?>
<RQ V="1"><DAT DI="14280" DT="0" FN="4001423294" TN="3301911428" V="1" ZN="0"><C T="0"><L N="1">ztm000355812</L><P N="2" C="300038" CD="22030756" NM="Ковбаса варена Молочна ДСТУ" Q="354" PRC="24400" SM="8638" TX="7"/><P N="3" C="300180" CD="22100017" NM="Сир Радомер" Q="204" PRC="40500" SM="8262" TX="7"/><P N="4" C="304619" CD="4820062054867" NM="Хліб Тостовий Молочний РІЗ" Q="1000" PRC="3450" SM="3450" TX="7"/><M N="5" T="1" NM="Картка" VCHASNOID="2" RRN="114463230641" PSNM="VISA" PA="S1ZT03AJ" PB="S1ZT03AJ" PC="Оплата" PD="4149XXXXXXXX9043" PE="598504" SM="20350"/><E N="6" NO="231" SM="20350" RND="1" TS="20260715185530" FN="4001423294"><TX TX="7" TXPR="0.00" TXSM="0" DTPR="0.00" DTSM="0" TXTY="0" TXAL="2"/></E></C><TS>20260715185530</TS></DAT></RQ>`)

	receipt, err := parser.Parse(raw)
	assert.NoError(t, err)
	if !assert.NotNil(t, receipt) {
		return
	}

	assert.Equal(t, int64(20350), receipt.Total)
	assert.Equal(t, "231", receipt.ReceiptNumber)
	if assert.Len(t, receipt.Items, 3) {
		assert.Equal(t, "Хліб Тостовий Молочний РІЗ", receipt.Items[2].Name)
		assert.Equal(t, int64(1), receipt.Items[2].Quantity)
		assert.Equal(t, int64(3450), receipt.Items[2].TotalAmount)
	}
	if assert.Len(t, receipt.Payments, 1) {
		assert.Equal(t, "VISA", receipt.Payments[0].Provider)
		assert.Equal(t, "4149XXXXXXXX9043", receipt.Payments[0].Mask)
	}
	kyiv, err := time.LoadLocation("Europe/Kyiv")
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, time.July, 15, 18, 55, 30, 0, kyiv).UnixMilli(), receipt.ReceiptDate.UnixMilli())
}

func TestSilpoXMLParser_ParseTipsSamples(t *testing.T) {
	parser := NewSilpoXMLParser()

	testCases := []struct {
		filename      string
		receiptNumber string
		total         int64
		itemCount     int
	}{
		{
			filename:      "dQamxBIKaeg.xml",
			receiptNumber: "231",
			total:         20350,
			itemCount:     3,
		},
		{
			filename:      "W9bEBMTh9_4.xml",
			receiptNumber: "75",
			total:         17880,
			itemCount:     3,
		},
		{
			filename:      "4860302367.xml",
			receiptNumber: "264392",
			total:         9200,
			itemCount:     3,
		},
		{
			filename:      "4980629892.xml",
			receiptNumber: "20047",
			total:         649700,
			itemCount:     1,
		},
		{
			filename:      "5085657832.xml",
			receiptNumber: "24499",
			total:         714600,
			itemCount:     3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "..", "tips", tc.filename))
			require.NoError(t, err)

			receipt, err := parser.Parse(raw)
			require.NoError(t, err)
			require.NotNil(t, receipt)

			assert.Equal(t, tc.receiptNumber, receipt.ReceiptNumber)
			assert.Equal(t, tc.total, receipt.Total)
			assert.Len(t, receipt.Items, tc.itemCount)
		})
	}
}

func TestSilpoXMLParser_ParseInvalidThenValid(t *testing.T) {
	parser := NewSilpoXMLParser()

	_, err := parser.Parse([]byte(`<broken xml`))
	assert.Error(t, err)

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "tips", "dQamxBIKaeg.xml"))
	require.NoError(t, err)

	receipt, err := parser.Parse(raw)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, "231", receipt.ReceiptNumber)
}

func TestSilpoXMLParser_FillsMissingPricePerUnitFromTotal(t *testing.T) {
	parser := NewSilpoXMLParser()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "tips", "676217.xml"))
	require.NoError(t, err)

	receipt, err := parser.Parse(raw)
	require.NoError(t, err)
	require.NotNil(t, receipt)

	byName := make(map[string]ParsedReceiptItem, len(receipt.Items))
	for _, item := range receipt.Items {
		byName[item.Name] = item
	}

	require.Contains(t, byName, "Кетчуп 380 г Торчин Лагідний д/пак")
	assert.Equal(t, int64(1), byName["Кетчуп 380 г Торчин Лагідний д/пак"].Quantity)
	assert.Equal(t, int64(3090), byName["Кетчуп 380 г Торчин Лагідний д/пак"].PricePerUnit)

	require.Contains(t, byName, "Засіб миючий для посуду 1 л Fairy Зелене Яблуко/")
	assert.Equal(t, int64(1), byName["Засіб миючий для посуду 1 л Fairy Зелене Яблуко/"].Quantity)
	assert.Equal(t, int64(12890), byName["Засіб миючий для посуду 1 л Fairy Зелене Яблуко/"].PricePerUnit)

	require.Contains(t, byName, "Дезодорант 150 мл Rexona Пр.кристал/блакитний ма")
	assert.Equal(t, int64(1), byName["Дезодорант 150 мл Rexona Пр.кристал/блакитний ма"].Quantity)
	assert.Equal(t, int64(7490), byName["Дезодорант 150 мл Rexona Пр.кристал/блакитний ма"].PricePerUnit)
}
