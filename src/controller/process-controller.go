package controller

import (
	"os"
	"server/entity"
	"server/services"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	bucketName string = "depression-app-backend-audio"
)

// var (
// 	bucketName string = "depressed-project"
// )

type ProcessController interface {
	AsynchronousSTT(ctx *gin.Context, fileID string) (interface{}, error)
	VAD(token string, fileID string) error
	ClearFile(ctx *gin.Context) error
}

type processController struct {
	processService services.ProcessService
}

func NewProcessController() ProcessController {
	return &processController{
		processService: services.NewProcessService(),
	}
}

func (controller *processController) jwtSecret(token string) string {
	secret := strings.Split(token, ".")[2]
	return secret

}

func (controller *processController) AsynchronousSTT(ctx *gin.Context, fileID string) (interface{}, error) {
	var format entity.SttProcessFormat
	secret := controller.jwtSecret(ctx.Request.Header.Get("Authorization"))
	format.JwtSecret = secret
	err := controller.processService.SetupClient(&format)
	if err != nil {
		return nil, err

	}
	defer format.Client.Close()
	controller.processService.SetupBucket(&format, bucketName)
	result, err := controller.processService.AsynchronousSTT(&format, fileID)
	if err != nil {
		return nil, err

	}

	return result, nil

}

func (controller *processController) ClearFile(ctx *gin.Context) error {
	var format entity.SttProcessFormat
	secret := controller.jwtSecret(ctx.Request.Header.Get("Authorization"))
	format.JwtSecret = secret
	err := controller.processService.SetupClient(&format)
	if err != nil {
		return err

	}
	defer format.Client.Close()
	controller.processService.SetupBucket(&format, bucketName)
	err = controller.processService.ClearFile(&format)
	return err
}

// func (controller *processController) VAD(token string) interface{} {
func (controller *processController) VAD(token string, fileID string) error {
	secret := controller.jwtSecret(token)
	path, _ := os.Getwd()

	tran_dir := path + "/transition-folder"
	os.Mkdir(tran_dir, os.ModePerm)
	os.Chmod(tran_dir, os.ModePerm)
	in_tran_dir := path + "/transition-folder" + "/" + secret
	os.Mkdir(in_tran_dir, os.ModePerm)
	os.Chmod(in_tran_dir, os.ModePerm)

	output_dir := path + "/output-folder"
	os.Mkdir(output_dir, os.ModePerm)
	os.Chmod(output_dir, os.ModePerm)
	in_output_dir := path + "/output-folder" + "/" + secret
	os.Mkdir(in_output_dir, os.ModePerm)
	os.Chmod(in_output_dir, os.ModePerm)

	result_dir := path + "/result-folder"
	os.Mkdir(result_dir, os.ModePerm)
	os.Chmod(result_dir, os.ModePerm)
	in_result_dir := path + "/result-folder" + "/" + secret
	os.Mkdir(in_result_dir, os.ModePerm)
	os.Chmod(in_result_dir, os.ModePerm)

	err := controller.processService.VAD(secret, fileID)
	if err != nil {
		return err
	}
	return nil
}
