package upload

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/froggy-12/mooshroombase_v2/types"
	"github.com/froggy-12/mooshroombase_v2/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func HandleUploadImageFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	file := form.File["file"][0]

	if file == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "fo files found"})
	}

	if !utils.IsImage(file.Filename) {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Only Image files can be accepted on this route"})
	}

	uploadsDir := filepath.Join(".", "uploads/images")
	err = os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		log.Println("Error creating folder: ", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Internal server error"})
	}

	// Generate a UUID for the filename
	uuidFilename := uuid.New().String() + filepath.Ext(file.Filename)

	err = c.SaveFile(file, filepath.Join(uploadsDir, uuidFilename))
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload file"})
	}

	return c.JSON(types.SingleFileUploadedSuccessResponse{FileName: uuidFilename, Message: "File Upload Successfull"})
}

func HandleUploadMultipleImageFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	files := form.File["file"]

	if len(files) == 0 {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "fo files found"})
	}

	var uploadedFiles []string
	var tempFiles []string

	uploadsDir := filepath.Join(".", "uploads/images")
	err = os.MkdirAll(uploadsDir, os.ModePerm)
	if err != nil {
		log.Println("Error creating folder: ", err.Error())
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Internal server error"})
	}

	// Create a temporary folder for storing files
	tempDir, err := os.MkdirTemp(filepath.Dir(uploadsDir), "temp-")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to create temporary folder"})
	}
	defer os.RemoveAll(tempDir)

	for _, file := range files {
		if !utils.IsImage(file.Filename) {
			for _, tempFile := range tempFiles {
				os.Remove(tempFile)
			}
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Only Image files can be accepted on this route"})
		}

		uuidFilename := uuid.New().String() + filepath.Ext(file.Filename)
		tempFile := filepath.Join(tempDir, uuidFilename)
		err = c.SaveFile(file, tempFile)
		if err != nil {
			for _, tempFile := range tempFiles {
				os.Remove(tempFile)
			}
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload image files"})
		}

		tempFiles = append(tempFiles, tempFile)
		uploadedFiles = append(uploadedFiles, uuidFilename)
	}

	// Move files from temporary folder to final storage location
	for _, tempFile := range tempFiles {
		err = os.Rename(tempFile, filepath.Join(uploadsDir, filepath.Base(tempFile)))
		if err != nil {
			log.Println("Error moving file to final storage location: ", err.Error())
			return c.Status(500).JSON(types.ErrorResponse{Error: "Failed to upload image files"})
		}
	}

	return c.JSON(types.MultipleFileUploadedSuccessResponse{FileNames: uploadedFiles, Message: "Image Files Upload Successfull"})
}

func HandleUploadSingleMusicFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	file := form.File["file"][0]

	if file == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "fo files found"})
	}

	if !utils.IsMusic(file.Filename) {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Only music files can be accepted on this route"})
	}

	uuidFilename, err := utils.UploadFile(c, file, "musics")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload music file"})
	}

	return c.JSON(types.SingleFileUploadedSuccessResponse{FileName: uuidFilename, Message: "Music File Upload Successfull"})
}

func HandleUploadMultipleMusicFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	files := form.File["file"]

	if len(files) == 0 {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "fo files found"})
	}

	var uploadedFiles []string
	var tempFiles []string

	for _, file := range files {
		if !utils.IsMusic(file.Filename) {
			for _, tempFile := range tempFiles {
				os.Remove(tempFile)
			}
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Only music files can be accepted on this route"})
		}

		uuidFilename, err := utils.UploadFile(c, file, "musics")
		if err != nil {
			for _, tempFile := range tempFiles {
				os.Remove(tempFile)
			}
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload music files"})
		}

		tempFiles = append(tempFiles, filepath.Join("uploads/musics", uuidFilename))
		uploadedFiles = append(uploadedFiles, uuidFilename)
	}

	return c.JSON(types.MultipleFileUploadedSuccessResponse{FileNames: uploadedFiles, Message: "Music Files Upload Successfull"})
}

func HandleUploadSingleVideoFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	file := form.File["file"][0]

	if file == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "fo files found"})
	}

	if !utils.IsVideo(file.Filename) {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Only video files can be accepted on this route"})
	}

	uuidFilename, err := utils.UploadFile(c, file, "videos")
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload video file"})
	}

	return c.JSON(types.SingleFileUploadedSuccessResponse{FileName: uuidFilename, Message: "Video File Upload Successfull"})
}

func HandleUploadMultiVideoFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	files := form.File["file"]

	if len(files) == 0 {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "fo files found"})
	}

	var uploadedFiles []string
	var tempFiles []string

	for _, file := range files {
		if !utils.IsVideo(file.Filename) {
			for _, tempFile := range tempFiles {
				os.Remove(tempFile)
			}
			return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Only video files can be accepted on this route"})
		}

		uuidFilename, err := utils.UploadFile(c, file, "videos")
		if err != nil {
			for _, tempFile := range tempFiles {
				os.Remove(tempFile)
			}
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload video files"})
		}

		tempFiles = append(tempFiles, filepath.Join("uploads/videos", uuidFilename))
		uploadedFiles = append(uploadedFiles, uuidFilename)
	}

	return c.JSON(types.MultipleFileUploadedSuccessResponse{FileNames: uploadedFiles, Message: "Video Files Upload Successfull"})
}

func HandleAnyFormatSingleFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	file := form.File["file"][0]

	if file == nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "fo files found"})
	}

	uuid := uuid.New()
	uuidFilename := fmt.Sprintf("%s%s", uuid, filepath.Ext(file.Filename))

	err = utils.UploadAnyFile(c, file, "files", uuidFilename)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload file"})
	}

	return c.JSON(types.SingleFileUploadedSuccessResponse{FileName: uuidFilename, Message: "File Upload Successfull"})
}

func HandleAnyFormatMultiFile(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "Invalid request it should be multipart form"})
	}
	files := form.File["file"]

	if len(files) == 0 {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "no files found"})
	}

	var uploadedFiles []string

	for _, file := range files {
		uuid := uuid.New()
		uuidFilename := fmt.Sprintf("%s%s", uuid, filepath.Ext(file.Filename))

		err = utils.UploadAnyFile(c, file, "files", uuidFilename)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to upload files"})
		}

		uploadedFiles = append(uploadedFiles, uuidFilename)
	}

	return c.JSON(types.MultipleFileUploadedSuccessResponse{FileNames: uploadedFiles, Message: "Files Upload Successfull"})
}

func HandleDeleteFile(c *fiber.Ctx) error {
	filename := c.Query("filename")
	folder := c.Query("folder")

	if filename == "" || folder == "" {
		return c.Status(http.StatusBadRequest).JSON(types.ErrorResponse{Error: "filename and folder are required"})
	}

	err := utils.DeleteFile(filename, folder)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(types.ErrorResponse{Error: "Failed to delete specified file. file might not existed"})
	}

	return c.JSON(types.DeleteSuccessResponse{Message: "File deleted successfully", FileName: filename})
}
