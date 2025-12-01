package handlers

import (
	"net/http"

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

    req := domain.FileUploadRequest{
        FileName:     c.FormValue("file_name"),
        Description:  c.FormValue("description"),
        Subject:      c.FormValue("subject"),
        GroupAllowed: c.FormValue("group_allowed"),
        Type:         c.FormValue("type"),
    }

    file, err := c.FormFile("file")
    if err != nil {
        return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
            Status: "error",
            Error:  "invalid file upload",
        })
    }

    src, err := file.Open()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
            Status: "error",
            Error:  "failed to open uploaded file",
        })
    }
    defer src.Close()

    path, err := h.FileService.UploadFileService(
        c.Request().Context(),
        file.Filename,
        src,
        req,
    )

    if err != nil {
        return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
            Status: "error",
            Error:  "failed to upload file: " + err.Error(),
        })
    }

    return c.JSON(http.StatusOK, domain.SuccessResponse{
        Status:  "success",
        Message: "file uploaded successfully",
        Data:    map[string]string{"file_url": path},
    })
}
