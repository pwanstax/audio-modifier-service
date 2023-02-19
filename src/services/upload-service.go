package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"server/entity"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type UploadService interface {
	Upload(uploadFormat entity.UploadFormat, fileID string) error
	JWTSecret(token string) string
}

type uploadService struct{}

func NewUploadService() UploadService {
	return &uploadService{}

}

const (
	projectID  = "indiv-summer-2020"            // FILL IN WITH YOURS
	bucketName = "depression-app-backend-audio" // FILL IN WITH YOURS
)

// const (
// 	projectID  = "speech-text-1625481141336" // FILL IN WITH YOURS
// 	bucketName = "depressed-project"         // FILL IN WITH YOURS
// )

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

var uploader *ClientUploader

func init() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "admin.json")
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	uploader = &ClientUploader{
		cl:         client,
		bucketName: bucketName,
		projectID:  projectID,
		uploadPath: "",
	}
}

func (clientUploader *ClientUploader) UploadFile(file *multipart.FileHeader, secret string, fileID string) error {
	suffix := strings.Split(file.Filename, ".")[len(strings.Split(file.Filename, "."))-1]
	fileName := fileID + "." + suffix
	// fileName := file.Filename
	readFile, err := file.Open()
	if err != nil {
		return err

	}
	defer readFile.Close()
	ctx := context.Background()
	clientUploader.uploadPath = secret + "/"
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	wc := clientUploader.cl.Bucket(clientUploader.bucketName).Object(clientUploader.uploadPath + fileName).NewWriter(ctx)
	if _, err := io.Copy(wc, readFile); err != nil {
		return fmt.Errorf("io.Copy: %v", err)

	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)

	}
	return nil

}

func (service *uploadService) JWTSecret(token string) string {
	stripString := strings.Split(token, ".")
	return stripString[2]

}

func (service *uploadService) Upload(uploadFormat entity.UploadFormat, fileID string) error {
	token := uploadFormat.Token
	secret := service.JWTSecret(token)

	er := service.UploadToLocal(uploadFormat, token, fileID)
	if er != nil {
		return er

	}

	err := uploader.UploadFile(uploadFormat.AudioFile, secret, fileID)
	if err != nil {
		return err

	}
	return nil

}

func (service *uploadService) UploadToLocal(uploadFormat entity.UploadFormat, token string, fileID string) error {

	file, err := uploadFormat.AudioFile.Open()
	if err != nil {
		return err
	}

	fileNameOriginal := uploadFormat.AudioFile.Filename
	fileName := ""
	for _, i := range fileNameOriginal {
		if (string(i) == " ") || (string(i) == "(") || (string(i) == ")") {
			fileName += ""
		} else {
			fileName += string(i)
		}
	}
	fileType := strings.Split(fileName, ".")[len(strings.Split(fileName, "."))-1]
	fileName_without_type := strings.Join(strings.Split(fileName, ".")[:len(strings.Split(fileName, "."))-1], ".")
	fileName_withKey := fileName_without_type + fileID + "." + fileType

	storagePath, err := service.storagePath()
	if err != nil {
		return err
	}

	savePath := storagePath + "/" + service.JWTSecret(token) + "/" + fileName_withKey

	saveFile, err := os.Create(savePath)
	if err != nil {
		return err
	}

	defer saveFile.Close()
	_, err = io.Copy(saveFile, file)
	if err != nil {
		return err
	}
	return nil

}

func (service *uploadService) storagePath() (string, error) {
	serverPath, err := os.Getwd()
	if err != nil {
		return "", err

	}
	storagePath := serverPath + "/storage"
	return storagePath, nil

}
