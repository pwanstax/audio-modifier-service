package services

import "mime/multipart"

type FeatureService interface {
	CanProcessVAD(file *multipart.FileHeader) string
	CanProcessSpeechtoText(file *multipart.FileHeader) string
}

type featureService struct {
}

func NewFeatureService() FeatureService {
	return &featureService{}
}

func (service *featureService) CanProcessVAD(file *multipart.FileHeader) string {
	status := "OK"
	return status

}

func (service *featureService) CanProcessSpeechtoText(file *multipart.FileHeader) string {
	status := "OK"
	return status

}
