package csv

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"path"
)

type Repository struct {
	folderPath string
}

func New(folderPath string) Repository {
	return Repository{folderPath: folderPath}
}

const extensionCSV = ".csv"

// Save creates csv file which stores logs in csv format and returns filename with saved logs
func (r Repository) Save(csv string) (string, error) {
	fileName := uuid.New().String() + extensionCSV
	filePath := path.Join(r.folderPath, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("error while creating: %w", err)
	}

	defer func() { _ = file.Close() }()

	_, err = file.WriteString(csv)
	if err != nil {
		return "", fmt.Errorf("error while writing: %w", err)
	}

	return fileName, nil
}
