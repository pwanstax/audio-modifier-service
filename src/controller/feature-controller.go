package controller

import (
	"server/services"

	"github.com/gin-gonic/gin"
)

type FeatureController interface {
	CheckFeature(ctx *gin.Context) map[string]string
}

type featureController struct {
	featureService services.FeatureService
}

func NewFeatureController() FeatureController {
	return &featureController{
		featureService: services.NewFeatureService(),
	}

}

// CheckFeature checks which feature is able to run. Return map which keys are feature's name and value is "ok" if runnable else error message.
func (controller *featureController) CheckFeature(ctx *gin.Context) map[string]string {
	featureStatus := make(map[string]string)
	file, _ := ctx.FormFile("audiofile")
	featureStatus["stt"] = controller.featureService.CanProcessSpeechtoText(file)
	featureStatus["vad"] = controller.featureService.CanProcessVAD(file)
	return featureStatus

}
