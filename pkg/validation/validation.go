package validation

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/keircn/karu/pkg/errors"
)

func ValidatePositiveInt(value string, fieldName string) (int, error) {
	if value == "" {
		return 0, errors.New(errors.ValidationError, fieldName+" cannot be empty")
	}

	num, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.Wrapf(err, errors.ValidationError, "invalid %s: must be a number", fieldName)
	}

	if num < 0 {
		return 0, errors.New(errors.ValidationError, fieldName+" must be positive")
	}

	return num, nil
}

func ValidateNonEmptyString(value, fieldName string) error {
	if value == "" {
		return errors.New(errors.ValidationError, fieldName+" cannot be empty")
	}
	return nil
}

func ValidateURL(url string) error {
	if url == "" {
		return errors.New(errors.ValidationError, "URL cannot be empty")
	}

	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlPattern.MatchString(url) {
		return errors.New(errors.ValidationError, "invalid URL format")
	}

	return nil
}

func ValidateFilePath(path string) error {
	if path == "" {
		return errors.New(errors.ValidationError, "file path cannot be empty")
	}

	if !filepath.IsAbs(path) {
		return errors.New(errors.ValidationError, "file path must be absolute")
	}

	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return errors.Wrapf(err, errors.ValidationError, "directory does not exist: %s", dir)
	}

	return nil
}

func EnsureDirectoryExists(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.Wrapf(err, errors.ValidationError, "failed to create directory: %s", path)
	}
	return nil
}

func CreateFile(outputPath string) (*os.File, error) {
	dir := filepath.Dir(outputPath)
	if err := EnsureDirectoryExists(dir); err != nil {
		return nil, err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return nil, errors.Wrapf(err, errors.ValidationError, "failed to create file: %s", outputPath)
	}

	return file, nil
}
