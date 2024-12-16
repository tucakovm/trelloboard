package repository

import (
	"encoding/base64"
	"fmt"
	"github.com/colinmarc/hdfs/v2"
	"log"
	"os"
)

type HDFSRepository struct {
	Client *hdfs.Client
	logger *log.Logger
}

func NewHDFSRepository(logger *log.Logger, namenodeURL string) (*HDFSRepository, error) {
	client, err := hdfs.New(namenodeURL)
	if err != nil {
		return nil, fmt.Errorf("error creating HDFS client: %v", err)
	}
	return &HDFSRepository{
		Client: client,
		logger: logger,
	}, nil
}
func (repo *HDFSRepository) Close() {
	if repo.Client != nil {
		repo.Client.Close()
	}
}

func (repo *HDFSRepository) UploadFile(taskID, fileName string, content string) error {
	log.Println("Started upload")
	log.Println("Task ID:", taskID)
	log.Println("File name:", fileName)
	log.Println("end of file name")
	log.Println("start file content")
	log.Println("File content:", content)
	log.Println("end of file content")
	// Ensure HDFS client is initialized
	// For some reason it is nil otherwise
	// FIX THIS
	if repo.Client == nil {
		log.Println("HDFS client is not initialized. Initializing...")
		client, err := hdfs.New("namenode:8020")
		if err != nil {
			repo.logger.Printf("Failed to initialize HDFS client: %v", err)
			return fmt.Errorf("failed to initialize HDFS client: %v", err)
		}
		repo.Client = client
		log.Println("HDFS client initialized successfully")
	}

	// Decode Base64 content
	decodedContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		repo.logger.Printf("Error decoding Base64 content: %v", err)
		return fmt.Errorf("failed to decode Base64 content: %v", err)
	}
	log.Printf("Base64 content decoded, size: %d bytes", len(decodedContent))

	hdfsPath := fmt.Sprintf("/tasks/%s/%s", taskID, fileName)

	if repo.Client == nil {
		log.Println("HDFS client is not initialized.")
	}
	file, err := repo.Client.Create(hdfsPath)
	if err != nil {
		return fmt.Errorf("failed to create file on HDFS at %s: %v", hdfsPath, err)
	}
	defer file.Close()

	_, err = file.Write(decodedContent)
	if err != nil {
		log.Println("failed to write to HDFS at %s: %v", hdfsPath, err)
		log.Println("file content", decodedContent)
		return fmt.Errorf("failed to write content to file on HDFS at %s: %v", hdfsPath, err)
	}
	log.Println("File written to HDFS")

	// Flush the file to ensure all data is written
	err = file.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush file to HDFS at %s: %v", hdfsPath, err)
	}
	log.Println("File flushed successfully")

	log.Printf("Successfully uploaded file %s to HDFS at %s\n", fileName, hdfsPath)
	return nil
}

func (repo *HDFSRepository) DownloadFile(taskID string, fileName string) ([]byte, error) {

	// Check if repo.client is nil
	if repo.Client == nil {
		repo.logger.Println("Error: repo.client is nil")
		return nil, fmt.Errorf("HDFS client is not initialized")
	}
	hdfsPath := "/tasks/" + taskID + "/" + fileName

	// Open file in HDFS
	file, err := repo.Client.Open(hdfsPath)
	if err != nil {
		log.Println("Error opening file in HDFS:", err)
		return nil, err
	}
	defer file.Close()

	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil {
		log.Println("Error reading file in HDFS:", err)
		return nil, err
	}
	log.Println("File uploaded")

	return buffer[:n], nil
}

func (repo *HDFSRepository) DeleteFile(taskID string, fileName string) error {
	// Check if repo.client is nil
	if repo.Client == nil {
		repo.logger.Println("Error: repo.client is nil")
		return fmt.Errorf("HDFS client is not initialized")
	}

	hdfsPath := "/tasks/" + taskID + "/" + fileName

	err := repo.Client.Remove(hdfsPath)
	if err != nil {
		repo.logger.Println("Error deleting file in HDFS:", err)
		return err
	}

	return nil
}

func (repo *HDFSRepository) GetFileNamesForTask(taskID string) ([]string, error) {
	repo.logger.Printf("Started fetching file names for task_id: %s", taskID)

	// Check if the client is initialized
	if repo.Client == nil {
		repo.logger.Println("Error: HDFS client is not initialized")
		return nil, fmt.Errorf("HDFS client is not initialized")
	}

	hdfsPath := fmt.Sprintf("/tasks/%s", taskID)

	// Walk through the directory and collect file names
	var fileNames []string
	callbackFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			repo.logger.Println("Error during walk:", err)
			return err
		}
		if !info.IsDir() {
			fileNames = append(fileNames, info.Name())
		}
		return nil
	}

	// Walk the directory on HDFS and fetch file names
	err := repo.Client.Walk(hdfsPath, callbackFunc)
	if err != nil {
		repo.logger.Printf("Error walking HDFS directory %s: %v", hdfsPath, err)
		return nil, fmt.Errorf("failed to walk HDFS directory %s: %v", hdfsPath, err)
	}

	repo.logger.Printf("Found %d files for task_id: %s", len(fileNames), taskID)

	return fileNames, nil
}
