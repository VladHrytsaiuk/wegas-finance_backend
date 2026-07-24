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

const maxStoredImageDimension = 1600

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

	// Receipt photos do not need the full camera resolution. Normalizing on the
	// server keeps uploads from every client in the same compact format.
	img = resizeImage(img, maxStoredImageDimension)
	if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 75}); err != nil {
		return "", err
	}

	// Повертаємо відносний шлях для БД
	return "/uploads/" + folder + "/" + filename, nil
}

func resizeImage(src image.Image, maxDimension int) image.Image {
	bounds := src.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= maxDimension && height <= maxDimension {
		return src
	}

	newWidth, newHeight := width, height
	if width >= height {
		newWidth = maxDimension
		newHeight = height * maxDimension / width
	} else {
		newHeight = maxDimension
		newWidth = width * maxDimension / height
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := range newHeight {
		sourceY := bounds.Min.Y + y*height/newHeight
		for x := range newWidth {
			sourceX := bounds.Min.X + x*width/newWidth
			dst.Set(x, y, src.At(sourceX, sourceY))
		}
	}

	return dst
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
