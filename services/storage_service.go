package services

import (
	"errors"
	"image"
	"image/jpeg"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// StorageService - інтерфейс для роботи зі сховищем файлів
type StorageService interface {
	SaveImage(fileHeader *multipart.FileHeader, folder string) (string, error)
	SaveFile(fileHeader *multipart.FileHeader, folder string) (string, error)
	DeleteFile(path string) error
}

type localStorageService struct {
	baseDir string
}

func NewLocalStorageService(baseDir string) StorageService {
	return &localStorageService{baseDir: baseDir}
}

func (s *localStorageService) SaveImage(fileHeader *multipart.FileHeader, folder string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	uploadDir := filepath.Join(s.baseDir, folder)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	// Зберігаємо як JPEG з унікальним ім'ям
	filename := uuid.NewString() + ".jpg"
	filePath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Декодуємо картинку
	img, _, err := image.Decode(file)
	if err != nil {
		return "", errors.New("invalid image format")
	}

	// Стиснення 80%
	if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 80}); err != nil {
		return "", err
	}

	// Повертаємо відносний шлях для БД
	return "/uploads/" + folder + "/" + filename, nil
}

func (s *localStorageService) SaveFile(fileHeader *multipart.FileHeader, folder string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	uploadDir := filepath.Join(s.baseDir, folder)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", err
	}

	originalExt := filepath.Ext(fileHeader.Filename)
	filename := uuid.NewString() + originalExt
	filePath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", err
	}

	return "/uploads/" + folder + "/" + filename, nil
}

func (s *localStorageService) DeleteFile(path string) error {
	if path == "" {
		return nil
	}
	// Шлях у БД починається з /uploads/, тому додаємо крапку для відносного шляху на диску
	// Використовуємо strings.TrimPrefix про всяк випадок
	cleanPath := filepath.Join(".", strings.TrimPrefix(path, "/"))
	return os.Remove(cleanPath)
}
