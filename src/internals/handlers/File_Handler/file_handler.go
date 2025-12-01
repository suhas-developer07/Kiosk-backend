package handlers

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"

	"github.com/suhas-developer07/Kiosk-backend/src/internals/service"
)

type FileHandler struct {
	FileService *service.FileService
	Validator   *validator.Validate
}

func NewFileHandler(fs *service.FileService) *FileHandler {
	return &FileHandler{
		FileService: fs,
		Validator: validator.New(),
	}
}

func (h *FileHandler) UploadFileHandler(c echo.Context) error {
	//Basic implementation of flow of file upload
	var req domain.FileUploadRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "invalid request body",
		})
	}

	if err := h.Validator.Struct(req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Some required details are missing or incorrect: " + strings.ReplaceAll(err.Error(), "\n", ", "),
		})
	}

	if err := h.FileService.UploadFileService(c.Request().Context(), req); err != nil {
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "failed to upload file: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "file uploaded successfully",
	})

}
