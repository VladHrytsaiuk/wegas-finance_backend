package services

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/VladHrytsaiuk/wegas-finance/backend/models"
	"github.com/VladHrytsaiuk/wegas-finance/backend/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestReceiptIngestionService_IngestXML(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	xmlPath := filepath.Join("..", "..", "tips", "dQamxBIKaeg.xml")
	rawXML, err := os.ReadFile(xmlPath)
	require.NoError(t, err)

	fileHeader := createMultipartFileHeader(t, "receipt.xml", rawXML)
	inbox := new(MockInboxService)
	user := &models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1"}

	inbox.On("Create", mock.MatchedBy(func(input InboxCreateInput) bool {
		assert.Equal(t, models.InboxEntryStatusNeedsAccount, input.Status)
		assert.Equal(t, models.ReceiptSourceTypeXML, input.SourceType)
		assert.Equal(t, models.ReceiptOriginTelegramXML, input.Origin)
		assert.Equal(t, "231", input.ReceiptNumber)
		assert.Equal(t, "UAH", input.Currency)
		if assert.NotNil(t, input.Total) {
			assert.Equal(t, int64(20350), *input.Total)
		}
		if assert.NotNil(t, input.Subtotal) {
			assert.Equal(t, int64(20350), *input.Subtotal)
		}
		assert.Len(t, input.Items, 3)
		return true
	}), user).Return(&models.InboxEntry{
		Base:   models.Base{ID: "inbox-1"},
		Status: models.InboxEntryStatusNeedsAccount,
	}, nil).Once()

	service := NewReceiptIngestionService(db, inbox)
	entry, err := service.IngestXML(fileHeader, user)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "inbox-1", entry.ID)
	inbox.AssertExpectations(t)
}

func TestReceiptIngestionService_IngestURL(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	xmlPath := filepath.Join("..", "..", "tips", "dQamxBIKaeg.xml")
	rawXML, err := os.ReadFile(xmlPath)
	require.NoError(t, err)

	inbox := new(MockInboxService)
	user := &models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1"}
	sourceURL := "https://receipt.silpo.elkasa.com.ua/demo"
	xmlURL := "https://receipt.silpo.elkasa.com.ua/receipt.xml"

	inbox.On("Create", mock.MatchedBy(func(input InboxCreateInput) bool {
		assert.Equal(t, models.InboxEntryStatusNeedsAccount, input.Status)
		assert.Equal(t, models.ReceiptSourceTypeURL, input.SourceType)
		assert.Equal(t, models.ReceiptOriginTelegramURL, input.Origin)
		assert.Equal(t, sourceURL, input.SourceURL)
		assert.Equal(t, "Сільпо", input.Merchant)
		assert.Len(t, input.Items, 3)
		return true
	}), user).Return(&models.InboxEntry{
		Base:   models.Base{ID: "inbox-2"},
		Status: models.InboxEntryStatusNeedsAccount,
	}, nil).Once()

	service := NewReceiptIngestionService(db, inbox)
	impl := service.(*receiptIngestionService)
	impl.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			switch req.URL.String() {
			case sourceURL:
				return httpResponse(http.StatusOK, "text/html; charset=utf-8", `<html><body><a href="`+xmlURL+`">xml</a></body></html>`), nil
			case xmlURL:
				return httpResponse(http.StatusOK, "application/xml", string(rawXML)), nil
			default:
				return httpResponse(http.StatusNotFound, "text/plain", "not found"), nil
			}
		}),
	}

	entry, err := service.IngestURL(sourceURL, user)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "inbox-2", entry.ID)
	inbox.AssertExpectations(t)
}

func TestReceiptIngestionService_IngestURL_FromReceiptHTML(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	inbox := new(MockInboxService)
	user := &models.User{Base: models.Base{ID: "user-1"}, FamilyID: "family-1"}
	sourceURL := "https://receipt.silpo.elkasa.com.ua/dKjBFHIl5s8"

	htmlBody := `<html><body>
<div class="centered">&#x422;&#x41E;&#x412; &quot;&#x421;&#x456;&#x43B;&#x44C;&#x43F;&#x43E;-&#x424;&#x443;&#x434;&quot;</div>
<table class="cheque-goods w-document">
  <tbody><tr><td>ШК 1</td></tr><tr><td>Сік1.0GalicЯблучП/о</td><td>59.99</td><td>A</td></tr></tbody>
  <tbody><tr><td>ШК 2</td></tr><tr><td>БананКг</td></tr><tr><td>0.812 X 76.90</td><td>62.44</td><td>A</td></tr></tbody>
</table>
<div># В т.ч. знижка 118.50</div>
<table><tr><td><b>ДО СПЛАТИ:</b></td><td><b>328.89 ГРН</b></td></tr></table>
<table><tr><td>ПЛАТІЖНА СИСТЕМА</td><td>MASTERCARD</td></tr><tr><td>ЕПЗ</td><td>XXXXXXXXXXXX6666</td></tr></table>
<table><tr><td>ЧЕК № :</td><td>dKjBFHIl5s8</td></tr><tr><td>ЧАС :</td><td>21:30:15 06.06.2026</td></tr></table>
</body></html>`

	inbox.On("Create", mock.MatchedBy(func(input InboxCreateInput) bool {
		assert.Equal(t, models.InboxEntryStatusNeedsAccount, input.Status)
		assert.Equal(t, models.ReceiptSourceTypeURL, input.SourceType)
		assert.Equal(t, models.ReceiptOriginTelegramURL, input.Origin)
		assert.Equal(t, sourceURL, input.SourceURL)
		assert.Equal(t, `ТОВ "Сільпо-Фуд"`, input.Merchant)
		assert.Equal(t, "dKjBFHIl5s8", input.ReceiptNumber)
		if assert.NotNil(t, input.Total) {
			assert.Equal(t, int64(32889), *input.Total)
		}
		assert.Len(t, input.Items, 2)
		return true
	}), user).Return(&models.InboxEntry{
		Base:   models.Base{ID: "inbox-html"},
		Status: models.InboxEntryStatusNeedsAccount,
	}, nil).Once()

	service := NewReceiptIngestionService(db, inbox)
	impl := service.(*receiptIngestionService)
	impl.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == sourceURL {
				return httpResponse(http.StatusOK, "text/html; charset=utf-8", htmlBody), nil
			}
			return httpResponse(http.StatusNotFound, "text/plain", "not found"), nil
		}),
	}

	entry, err := service.IngestURL(sourceURL, user)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, "inbox-html", entry.ID)
	inbox.AssertExpectations(t)
}

func TestReceiptIngestionService_IngestXML_DeduplicatesExistingInboxEntry(t *testing.T) {
	db, err := repositories.SetupTestDB()
	require.NoError(t, err)

	xmlPath := filepath.Join("..", "..", "tips", "dQamxBIKaeg.xml")
	rawXML, err := os.ReadFile(xmlPath)
	require.NoError(t, err)

	user := &models.User{
		Base:     models.Base{ID: "user-1"},
		FamilyID: "family-1",
	}
	require.NoError(t, db.Create(user).Error)

	receiptDate := int64(1784130930000)
	total := int64(20350)
	subtotal := int64(20350)
	now := int64(1784460000000)

	receiptSource := &models.ReceiptSource{
		Base:          models.Base{ID: "source-1", CreatedAt: now, UpdatedAt: now, IsSynced: true},
		FamilyID:      user.FamilyID,
		UserID:        user.ID,
		Origin:        models.ReceiptOriginTelegramXML,
		SourceType:    models.ReceiptSourceTypeXML,
		RawPayload:    string(rawXML),
		Merchant:      "Сільпо",
		ReceiptNumber: "231",
		ReceiptDate:   &receiptDate,
		Subtotal:      &subtotal,
		Total:         &total,
		Currency:      "UAH",
	}
	require.NoError(t, db.Create(receiptSource).Error)
	require.NoError(t, db.Create(&models.ReceiptSourceItem{
		Base:            models.Base{ID: "item-1", CreatedAt: now, UpdatedAt: now, IsSynced: true},
		ReceiptSourceID: receiptSource.ID,
		Name:            "Тестовий товар",
		Quantity:        1,
		PricePerUnit:    20350,
		TotalAmount:     20350,
	}).Error)

	existingEntry := &models.InboxEntry{
		Base:            models.Base{ID: "inbox-existing", CreatedAt: now, UpdatedAt: now, IsSynced: true},
		FamilyID:        user.FamilyID,
		UserID:          user.ID,
		ReceiptSourceID: receiptSource.ID,
		Status:          models.InboxEntryStatusNeedsLink,
		SourceType:      models.ReceiptSourceTypeXML,
		Merchant:        "Сільпо",
		Total:           &total,
		Currency:        "UAH",
		OccurredAt:      &receiptDate,
	}
	require.NoError(t, db.Create(existingEntry).Error)

	inbox := new(MockInboxService)
	service := NewReceiptIngestionService(db, inbox)

	entry, err := service.IngestXMLBytes(rawXML, user)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, existingEntry.ID, entry.ID)
	inbox.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func createMultipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, req.ParseMultipartForm(int64(len(content))+1024))

	file, header, err := req.FormFile("file")
	require.NoError(t, err)
	_ = file.Close()

	return header
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func httpResponse(statusCode int, contentType string, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header: http.Header{
			"Content-Type": []string{contentType},
		},
		Body: io.NopCloser(bytes.NewBufferString(body)),
	}
}
