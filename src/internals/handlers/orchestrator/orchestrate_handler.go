package orchestrator

import (
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/service/orchestrator"
	"go.uber.org/zap"
)

type OrchestrateHandler struct {
	OrchestrateService *orchestrator.OrchestrateService
	Logger             *zap.SugaredLogger
	validate           *validator.Validate
}

func NewOrchestrateHandler(OrchestrateService *orchestrator.OrchestrateService, Logger *zap.SugaredLogger) *OrchestrateHandler {
	return &OrchestrateHandler{
		OrchestrateService: OrchestrateService,
		Logger:             Logger,
		validate:           validator.New(),
	}
}

func (h *OrchestrateHandler) UploadFileHandler(c echo.Context) error {
	ctx := c.Request().Context()

	FacultyID := c.Get("faculty_id").(string)

	req := domain.FileUploadRequest{
		Title:         c.FormValue("title"),
		Description:   c.FormValue("description"),
		Grade:         c.FormValue("grade"),
		Subject:       c.FormValue("subject"),
		Category:      c.FormValue("category"),
		GroupAllowed:  c.FormValue("group_allowed"),
		FileType:      c.FormValue("type"),
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

	allowedGrades := map[string]bool{"1PUC": true, "2PUC": true}
	if !allowedGrades[req.Grade] {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "invalid grade; allowed values: 1PUC, 2PUC",
		})
	}

	path, err := h.OrchestrateService.UploadFileService(
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

func (h *OrchestrateHandler) GetRecentUploadedFilesHandler(c echo.Context) error {
	ctx := c.Request().Context()

	FacultyID := c.Get("faculty_id").(string)

	if FacultyID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Not getting faculty id from the middleware",
		})
	}

	files, err := h.OrchestrateService.GetRecentUploadedFilesByFacultyIDService(ctx, FacultyID)

	if err != nil {
		h.Logger.Errorf("Internal server Error |err=%v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "Error",
			Error:  "Internal Error while fetching recent files",
		})
	}

	if len(files) == 0 {
		return c.JSON(http.StatusOK, domain.SuccessResponse{
			Status:  "success",
			Message: "You are not sent any files",
			Data:    []domain.File{},
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "successfully fetched the recent uploaded files files",
		Data:    files,
	})
}
func (h *OrchestrateHandler) FileDeleteRequestHandler(c echo.Context) error {
	ctx := c.Request().Context()

	fileID := strings.TrimSpace(c.Param("file_id"))

	type payload struct {
		Reason string `json:"reason"`
	}

	var req payload

	if err:=c.Bind(&req);err!=nil{
      return c.JSON(http.StatusBadRequest,domain.ErrorResponse{
		Status: "error",
		Error: "Invalid body reason is required",
	  })
	}

	clientIP := c.RealIP()

	h.Logger.Infof("Delete file request | fileID=%s | IP=%s", fileID, clientIP)

	if fileID == "" {
		h.Logger.Warnf("Delete file request failed: missing fileID | IP=%s", clientIP)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "fileId is required",
		})
	}

	err := h.OrchestrateService.FileDeleteRequestService(ctx, fileID,req.Reason)
	if err != nil {

		h.Logger.Errorf("Delete file request failed | fileID=%s | error=%v", fileID, err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}
	h.Logger.Infof("Delete file request succeeded | fileID=%s | IP=%s", fileID, clientIP)

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Delete request submitted successfully",
	})
}
