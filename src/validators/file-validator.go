package validators

import (
	"mime/multipart"
	"strings"
)

type AudioFileTypeError struct {
	message string
}

func (err *AudioFileTypeError) Error() string {
	return err.message

}

func fileType(fileName string) string {
	stripString := strings.Split(fileName, ".")
	lastIndex := len(stripString) - 1
	return stripString[lastIndex]

}

func ValidateAudioFile(file *multipart.FileHeader) error {
	if file == nil {
		return &AudioFileTypeError{message: "No File Has Been Sent"}

	}
	fileName := file.Filename
	fileType := fileType(fileName)
	validFileTypes := []string{"flac", "wav"}
	for _, validFileType := range validFileTypes {
		if validFileType == fileType {
			return nil

		}
	}
	errorMessage := fileType + " is not valid."
	return &AudioFileTypeError{message: errorMessage}

}
