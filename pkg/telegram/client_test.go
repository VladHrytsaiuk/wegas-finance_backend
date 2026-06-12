package telegram

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_SendMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	// Save original URL and restore after test
	originalURL := tgBaseURL
	tgBaseURL = server.URL + "/bot%s/%s"
	defer func() { tgBaseURL = originalURL }()

	client := NewClient("token", "chatid")
	err := client.SendMessage("hello")

	assert.NoError(t, err)
}

func TestClient_SendPhoto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.Header.Get("Content-Type"), "multipart/form-data")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"ok":true}`)
	}))
	defer server.Close()

	originalURL := tgBaseURL
	tgBaseURL = server.URL + "/bot%s/%s"
	defer func() { tgBaseURL = originalURL }()

	client := NewClient("token", "chatid")
	err := client.SendPhoto("caption", "photo.jpg", []byte("fake-photo-data"))

	assert.NoError(t, err)
}
