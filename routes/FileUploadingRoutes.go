package routes

import (
	"github.com/froggy-12/mooshroombase_v2/services/upload"
	"github.com/gofiber/fiber/v2"
)

func FileUploadingRoutes(router fiber.Router) {
	router.Post("/upload/image/single", upload.HandleUploadImageFile)
	router.Post("/upload/image/multi", upload.HandleUploadMultipleImageFile)
	router.Post("/upload/music/single", upload.HandleUploadSingleMusicFile)
	router.Post("/upload/music/multi", upload.HandleUploadMultipleMusicFile)
	router.Post("/upload/video/single", upload.HandleUploadSingleVideoFile)
	router.Post("/upload/video/multi", upload.HandleUploadMultiVideoFile)
	router.Post("/upload/video/multi", upload.HandleUploadMultiVideoFile)
	router.Post("/upload/any/single", upload.HandleAnyFormatSingleFile)
	router.Post("/upload/any/multi", upload.HandleAnyFormatMultiFile)
	router.Delete("/deletefile", upload.HandleDeleteFile)
}
