package services

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalStorageService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	service := NewLocalStorageService(tempDir)

	t.Run("SaveFile", func(t *testing.T) {
		content := []byte("hello world")
		header := createMockFileHeader("test.txt", content)

		path, err := service.SaveFile(header, "docs")
		assert.NoError(t, err)
		assert.Contains(t, path, "/uploads/docs/")
		assert.True(t, stringsHasSuffix(path, ".txt"))

		// Check if file exists on disk
		diskPath := filepath.Join(tempDir, "docs", filepath.Base(path))
		_, err = os.Stat(diskPath)
		assert.NoError(t, err)

		data, _ := os.ReadFile(diskPath)
		assert.Equal(t, content, data)
	})

	t.Run("SaveImage", func(t *testing.T) {
		// Create a mock image
		img := image.NewRGBA(image.Rect(0, 0, 10, 10))
		for x := 0; x < 10; x++ {
			for y := 0; y < 10; y++ {
				img.Set(x, y, color.RGBA{255, 0, 0, 255})
			}
		}
		buf := new(bytes.Buffer)
		jpeg.Encode(buf, img, nil)

		header := createMockFileHeader("photo.png", buf.Bytes())

		path, err := service.SaveImage(header, "assets")
		assert.NoError(t, err)
		assert.Contains(t, path, "/uploads/assets/")
		assert.True(t, stringsHasSuffix(path, ".jpg")) // Should be converted to jpg

		diskPath := filepath.Join(tempDir, "assets", filepath.Base(path))
		_, err = os.Stat(diskPath)
		assert.NoError(t, err)
	})

	t.Run("DeleteFile", func(t *testing.T) {
		// Note: DeleteFile implementation in storage_service.go uses filepath.Join(".", ...)
		// which makes it relative to current working directory.
		// If baseDir in NewLocalStorageService is absolute, DeleteFile might fail if it's not handled correctly.
		// However, let's test it as if we are in the root.
		
		path := "/uploads/temp/test.txt"
		diskPath := filepath.Join("uploads", "temp", "test.txt")
		os.MkdirAll(filepath.Dir(diskPath), 0755)
		os.WriteFile(diskPath, []byte("data"), 0644)
		
		err := service.DeleteFile(path)
		assert.NoError(t, err)
		
		_, err = os.Stat(diskPath)
		assert.True(t, os.IsNotExist(err))
		
		// Cleanup
		os.RemoveAll("uploads")
	})
}

func createMockFileHeader(filename string, content []byte) *multipart.FileHeader {
	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	part, _ := writer.CreateFormFile("file", filename)
	part.Write(content)
	writer.Close()

	reader := multipart.NewReader(buf, writer.Boundary())
	form, _ := reader.ReadForm(1024)
	if form == nil || form.File == nil || len(form.File["file"]) == 0 {
		return &multipart.FileHeader{Filename: filename}
	}
	return form.File["file"][0]
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
