package repository

import (
	"github.com/colinmarc/hdfs/v2"
	"log"
)

type HDFSRepository struct {
	client *hdfs.Client
	logger *log.Logger
}

func NewHDFSRepository(logger *log.Logger, hdfsUri string) (*HDFSRepository, error) {
	client, err := hdfs.New(hdfsUri)
	if err != nil {
		logger.Println("Error connecting to HDFS:", err)
		return nil, err
	}

	return &HDFSRepository{
		client: client,
		logger: logger,
	}, nil
}

func (repo *HDFSRepository) Close() {
	repo.client.Close()
}

func (repo *HDFSRepository) UploadFile(taskID string, fileName string, fileContent []byte) error {
	hdfsPath := "/tasks/" + taskID + "/" + fileName

	// Create directories if they don't exist
	err := repo.client.MkdirAll("/tasks/"+taskID, 0755)
	if err != nil {
		repo.logger.Println("Error creating directories:", err)
		return err
	}

	// Create file in HDFS
	file, err := repo.client.Create(hdfsPath)
	if err != nil {
		repo.logger.Println("Error creating file in HDFS:", err)
		return err
	}
	defer file.Close()

	_, err = file.Write(fileContent)
	if err != nil {
		repo.logger.Println("Error writing to file:", err)
		return err
	}

	return nil
}

func (repo *HDFSRepository) DownloadFile(taskID string, fileName string) ([]byte, error) {
	hdfsPath := "/tasks/" + taskID + "/" + fileName

	file, err := repo.client.Open(hdfsPath)
	if err != nil {
		repo.logger.Println("Error opening file in HDFS:", err)
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		repo.logger.Println("Error reading file in HDFS:", err)
		return nil, err
	}

	return buffer[:n], nil
}

func (repo *HDFSRepository) DeleteFile(taskID string, fileName string) error {
	hdfsPath := "/tasks/" + taskID + "/" + fileName

	err := repo.client.Remove(hdfsPath)
	if err != nil {
		repo.logger.Println("Error deleting file in HDFS:", err)
		return err
	}

	return nil
}
