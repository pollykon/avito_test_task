package csv

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"io"
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

func (r Repository) Get(name string) (string, error) {
	filePath := path.Join(r.folderPath, name+extensionCSV)

	file, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", ErrFileNotExist
		}
		return "", fmt.Errorf("error while opening: %w", err)
	}

	defer func() { _ = file.Close() }()

	var csv []byte
	var csvLen int

	for {
		csvLen, err = file.Read(csv)
		if err == io.EOF {
			break
		}
	}

	return string(csv[:csvLen]), nil
}
