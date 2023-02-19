package controller

import (
	"os"
	"server/entity"
	"server/services"
	"server/validators"

	"github.com/gin-gonic/gin"
)

type UploadController interface {
	Upload(ctx *gin.Context) (string, error)
	CreateSecret(token string) string
}

type uploadController struct {
	uploadService services.UploadService
}

func NewUploadController() UploadController {
	return &uploadController{
		uploadService: services.NewUploadService(),
	}
}

func (controller *uploadController) CreateUserFolder(token string) {
	secret := controller.uploadService.JWTSecret(token)
	userDir := "./storage" + "/" + secret
	os.Mkdir(userDir, os.ModePerm)
	os.Chmod(userDir, os.ModePerm)

}

func (controller *uploadController) CreateSecret(token string) string {
	return controller.uploadService.JWTSecret(token)
}

func (controller *uploadController) Upload(ctx *gin.Context) (string, error) {

	file, err := ctx.FormFile("audiofile")
	if err != nil {
		return "error", err

	}

	err = validators.ValidateAudioFile(file)
	if err != nil {
		return "error", err

	}

	token := ctx.Request.Header.Get("Authorization")

	controller.CreateUserFolder(token)
	fileID := ctx.GetHeader("amount")
	// tokenString := ctx.GetHeader("Authorization")
	// secret := strings.Split(tokenString, ".")

	//generate ID

	// files, err := ioutil.ReadDir("storage/" + secret[2])
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// temp := len(files) + 1

	// fileID := strconv.Itoa(temp)

	for len(fileID) < 9 {
		fileID = "0" + fileID
	}
	uploadFormat := entity.UploadFormat{
		AudioFile: file,
		Token:     token,
	}
	err = controller.uploadService.Upload(uploadFormat, fileID)
	if err != nil {
		return "error", err

	}
	return fileID, nil

}
