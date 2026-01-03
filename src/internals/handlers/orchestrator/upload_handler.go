package orchestrator

import (
	"net/http"

	"github.com/labstack/echo/v4"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/service/orchestrator"
	"go.uber.org/zap"
)

type UploadHandler struct {
	UploadService *orchestrator.UploadService
	Logger         *zap.SugaredLogger
}

func NewUploadHandler(UploadService *orchestrator.UploadService,Logger *zap.SugaredLogger) *UploadHandler {
	return &UploadHandler{
		UploadService: UploadService,
		Logger:         Logger,
	}
}

func (h *UploadHandler) UploadFileHandler(c echo.Context) error {
	ctx := c.Request().Context()

	FacultyID := c.Get("faculty_id").(string)

	req := domain.FileUploadRequest{
		Title:        c.FormValue("title"),
		Description:  c.FormValue("description"),
		Grade:        c.FormValue("grade"),
		Subject:      c.FormValue("subject"),
		Category:     c.FormValue("category"),
		GroupAllowed: c.FormValue("group_allowed"),
		FileType:     c.FormValue("type"),
	}

	//TODO : faculty id and faculty name comes from middleware -->done  need to test

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

	path, err := h.UploadService.UploadFileService(
		ctx,
		file.Filename,
		src,
		req,
		FacultyID,
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
