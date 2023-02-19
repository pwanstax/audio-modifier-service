package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"server/controller"
	"server/middlewares"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	tokenController  controller.TokenController  = controller.NewTokenController()
	uploadController controller.UploadController = controller.NewUploadController()
	// featureController controller.FeatureController = controller.NewFeatureController()
	processController controller.ProcessController = controller.NewProcessController()
)

func main() {

	setupServer() // don't forget to close DEBUG MODE BEFORE DEPLOYING
	// and don't forget to change language to TH

	server := gin.Default()
	server.Use(middlewares.CORSMiddleware(), middlewares.Logger())

	tokenRoute := server.Group("/token")
	{
		tokenRoute.GET("/getToken", func(ctx *gin.Context) {
			token := tokenController.GenerateToken()
			if token != "" {
				ctx.JSON(http.StatusOK, gin.H{
					"token": token,
				})
			} else {
				ctx.JSON(http.StatusUnauthorized, nil)
			}
		})
	}

	uploadRoute := server.Group("/upload", middlewares.AuthorizeJWT())
	{
		uploadRoute.POST("/audiofile", func(ctx *gin.Context) {
			if fileID, err := uploadController.Upload(ctx); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			} else {
				ctx.JSON(http.StatusOK, gin.H{
					"id": fileID,
				})
			}
		})
	}

	processRoute := server.Group("/process", middlewares.AuthorizeJWT())
	{
		processRoute.GET("/stt", func(ctx *gin.Context) {

			FileID := ctx.GetHeader("id")

			result, err := processController.AsynchronousSTT(ctx, FileID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			}

			// value := gjson.Parse(result).Get("stt")
			ctx.JSON(http.StatusOK, gin.H{
				"stt": result,
			})

		})
		processRoute.GET("/vad", func(ctx *gin.Context) {

			tokenString := ctx.GetHeader("Authorization")
			FileID := ctx.GetHeader("id")

			err := processController.VAD(tokenString, FileID)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
			}

			secret := strings.Split(tokenString, ".")
			files, err := ioutil.ReadDir("result-folder/" + secret[2])
			if err != nil {
				log.Fatal(err)
			}

			for _, f := range files {
				temp := strings.Join(strings.Split(f.Name(), ".")[:len(strings.Split(f.Name(), "."))-1], ".")
				fmt.Println(f.Name())
				if len(temp) > 9 {
					if temp[len(temp)-9:] == FileID {
						ctx.Header("Content-Type", "text/csv")
						os.Rename("result-folder/"+secret[2]+"/"+f.Name(), "result-folder/"+secret[2]+"/"+temp[:len(temp)-9]+".csv")
						ctx.File("result-folder/" + secret[2] + "/" + temp[:len(temp)-9] + ".csv")
						// os.Remove("result-folder/" + secret[2] + "/" + temp[:len(temp)-9] + ".csv")
						return
					}
				}
			}
			ctx.JSON(http.StatusOK, gin.H{
				"massage": "invalid id",
			})

		})
	}

	server.DELETE("/clear", func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		secret := strings.Split(tokenString, ".")[2]
		os.RemoveAll("./storage/" + secret)
		os.RemoveAll("./transition-folder/" + secret)
		os.RemoveAll("./output-folder/" + secret)
		os.RemoveAll("./result-folder/" + secret)
		processController.ClearFile(ctx)
	})

	port := envPort()

	server.Run(port)

}
