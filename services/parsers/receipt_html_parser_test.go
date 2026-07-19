package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSilpoHTMLParser_Parse(t *testing.T) {
	parser := NewSilpoHTMLParser()

	raw := []byte(`
<html><body>
<div class="centered">&#x422;&#x41E;&#x412; &quot;&#x421;&#x456;&#x43B;&#x44C;&#x43F;&#x43E;-&#x424;&#x443;&#x434;&quot;</div>
<table class="cheque-goods w-document">
  <tbody>
    <tr><td class="cheque-row-lcolumn" colspan="3">&#x428;&#x41A; 4820209560077</td></tr>
    <tr><td class="cheque-row-lcolumn no-break">Сік1.0GalicЯблучП/о</td><td class="cheque-row-rcolumn">59.99</td><td class="cheque-row-rcolumn">A</td></tr>
  </tbody>
  <tbody>
    <tr><td class="cheque-row-lcolumn" colspan="3">&#x428;&#x41A; 2732485</td></tr>
    <tr><td class="cheque-row-lcolumn no-break" colspan="3" style="padding-right: 0;">БананКг</td></tr>
    <tr><td class="cheque-row-lcolumn no-break">0.812 X 76.90</td><td class="cheque-row-rcolumn">62.44</td><td class="cheque-row-rcolumn">A</td></tr>
  </tbody>
</table>
<div class='borders-rl padding-rl w-document'># В т.ч. знижка 118.50</div>
<table class="w-document"><tr><td class='cheque-row-lcolumn'><b>ДО СПЛАТИ:</b></td><td class="cheque-row-rcolumn"><b>328.89 ГРН</b></td></tr></table>
<table class="w-document"><tr><td class="cheque-row-lcolumn">ПЛАТІЖНА СИСТЕМА</td><td class="cheque-row-rcolumn">MASTERCARD</td></tr><tr><td class='cheque-row-lcolumn'>ЕПЗ</td><td class="cheque-row-rcolumn">XXXXXXXXXXXX6666</td></tr></table>
<table class="w-document"><tbody><tr><td class='device-info-line-item'>ЧЕК № :</td><td class="device-info-line-item">dKjBFHIl5s8</td></tr><tr><td class='device-info-line-item'>ЧАС :</td><td class="device-info-line-item">21:30:15 06.06.2026</td></tr></tbody></table>
</body></html>`)
	receipt, err := parser.Parse(raw)
	require.NoError(t, err)
	require.NotNil(t, receipt)

	assert.Equal(t, `ТОВ "Сільпо-Фуд"`, receipt.Merchant)
	assert.Equal(t, "dKjBFHIl5s8", receipt.ReceiptNumber)
	assert.Equal(t, int64(32889), receipt.Total)
	assert.Equal(t, int64(11850), receipt.DiscountTotal)
	if assert.Len(t, receipt.Items, 2) {
		assert.Equal(t, "Сік1.0GalicЯблучП/о", receipt.Items[0].Name)
		assert.Equal(t, int64(5999), receipt.Items[0].TotalAmount)
		assert.Equal(t, "БананКг (0.812X76.90)", receipt.Items[1].Name)
		assert.Equal(t, int64(6244), receipt.Items[1].TotalAmount)
	}
	if assert.Len(t, receipt.Payments, 1) {
		assert.Equal(t, "MASTERCARD", receipt.Payments[0].Provider)
		assert.Equal(t, "XXXXXXXXXXXX6666", receipt.Payments[0].Mask)
	}
}
