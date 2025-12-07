package handlers

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/Files"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"

	"github.com/suhas-developer07/Kiosk-backend/src/internals/service"
)

type FileHandler struct {
	FileService *service.FileService
	Logger      *zap.SugaredLogger
}

func NewFileHandler(fs *service.FileService,Logger *zap.SugaredLogger) *FileHandler {
	return &FileHandler{
		FileService: fs,
        Logger:Logger,
	}
}

func (h *FileHandler) UploadFileHandler(c echo.Context) error {

	req := domain.FileUploadRequest{
		Title:        c.FormValue("title"),
		Description:  c.FormValue("description"),
		Grade:        c.FormValue("grade"),
		Subject:      c.FormValue("subject"),
		Category:     c.FormValue("category"),
		GroupAllowed: c.FormValue("group_allowed"),
		FileType:     c.FormValue("type"),
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

func (h *FileHandler) GetFilesByGradeAndSubjectHandler(c echo.Context) error {
	ctx := c.Request().Context()

	grade := strings.TrimSpace(strings.ToUpper(c.Param("grade")))
	subject := strings.TrimSpace(strings.Title(c.Param("subject")))

	if grade == "" || subject == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "grade and subject are required fields",
		})
	}

	allowedGrades := map[string]bool{"1PUC": true, "2PUC": true}
	if !allowedGrades[grade] {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "invalid grade; allowed values: 1PUC, 2PUC",
		})
	}

	// TODO:subject list can be loaded from config
	if subject == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "subject cannot be empty",
		})
	}

	h.Logger.Infof("Fetching files | Grade=%s | Subject=%s | IP=%s",
		grade, subject, c.RealIP(),
	)

	files, err := h.FileService.GetFileByGradeAndSubjectService(ctx, grade, subject)
	if err != nil {
		h.Logger.Errorf("Failed to fetch files | Grade=%s | Subject=%s | Error=%v",
			grade, subject, err,
		)

		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "internal error fetching files",
		})
	}

	if len(files) == 0 {
		return c.JSON(http.StatusOK, domain.SuccessResponse{
			Status:  "success",
			Message: "no files available for selected grade and subject",
			Data:    []domain.File{},
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "files fetched successfully",
		Data:    files,
	})
}

func (h *FileHandler) PrintUploadHandler(c echo.Context)error{
	ctx := c.Request().Context()

	var payload domain.PrintJobPayload

	if err := c.Bind(&payload);err!=nil{
		h.Logger.Warnf("Invalid print payload | IP=%s |Error=%v",c.RealIP(),err)
		return c.JSON(http.StatusBadRequest,domain.ErrorResponse{
			Status: "error",
			Error:"Invalid request body. Please ensure the JSON structure matches backend expectaions",
		})
	}

	if err := utils.ValidatePrintJobPayload(payload);err!=nil{
		h.Logger.Warnf("validation failed for printJob  | payload= %v |Error = %v",payload,err)
		return c.JSON(http.StatusBadRequest,domain.ErrorResponse{
			Status: "error",
			Error: err.Error(),
		})
	}
	h.Logger.Infof("Recieved print Job | FileID=%s | Copies=%d | IP=%s",
	        payload.FileID.Hex(),payload.Copies,c.RealIP(),
		)
	
	token,err := h.FileService.CreatePrintJobService(ctx,payload)
	if err != nil {
		h.Logger.Error("failed to create an printJob | payload= %v |Error |%v",payload,err)
		return c.JSON(http.StatusInternalServerError,domain.ErrorResponse{
			Status: "error",
			Error:"Internal error creating print job",
		})
	}

	return c.JSON(http.StatusOK,domain.SuccessResponse{
		Status: "success",
		Message: "Print Job created successfully",
		Data: token,
	})
}