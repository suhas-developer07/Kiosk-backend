package orchestrator

import (
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	model "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/faculties"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	"github.com/suhas-developer07/Kiosk-backend/src/internals/service/orchestrator"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
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
		validate: validator.New(),
	}
}

func (h *OrchestrateHandler) UploadFileHandler(c echo.Context) error {
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

func (h *OrchestrateHandler) AddFacultyHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var req model.FacultyPayload

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "Invalid Body",
		})
	}

	if err := h.validate.Struct(&req); err != nil {
		msg := utils.FormatValidationError(err)
		h.Logger.Warnf("Account validation failed | payload=%v | error=%v", req, msg)

		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	err := h.OrchestrateService.AddFacultyService(ctx, req)

	if err != nil {
		switch {
		case errors.Is(err, model.ErrEmailAlreadyExists):
			h.Logger.Warnf("Email already exists | email=%s", req.Email)
			return c.JSON(http.StatusConflict, domain.ErrorResponse{
				Status: "error",
				Error:  "Email already exists.",
			})

		default:
			h.Logger.Errorf("Failed to Add the faculty | email=%s | error=%v", req.Email, err)
			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "Internal server error",
			})
		}
	}
	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Faculty successfully Added ",
	})
}
