package filehandlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	apperrors "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/errors"
	domain "github.com/suhas-developer07/Kiosk-backend/src/internals/domain/files"
	service "github.com/suhas-developer07/Kiosk-backend/src/internals/service/file_service"
	"github.com/suhas-developer07/Kiosk-backend/src/pkg/utils"
	"go.uber.org/zap"
)

type FileHandler struct {
	FileService *service.FileService
	Logger      *zap.SugaredLogger
}

func NewFileHandler(fs *service.FileService, Logger *zap.SugaredLogger) *FileHandler {
	return &FileHandler{
		FileService: fs,
		Logger:      Logger,
	}
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

		if errors.Is(err, apperrors.ErrInvalidSubject) {
			h.Logger.Warnf("Invalid subject | Grade=%s |subject=%s", grade, subject)

			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Status: "error",
				Error:  "Invalid subject name:" + subject,
			})
		}
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

func (h *FileHandler) AccessFileHandler(c echo.Context) error {
	ctx := c.Request().Context()

	fileID := strings.TrimSpace(c.Param("file_id"))

	h.Logger.Infow(
		"access file request received",
		"file_id", fileID,
	)

	if fileID == "" {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  "file_id is required",
		})
	}

	signedURL, err := h.FileService.AccessFileService(ctx, fileID)
	if err != nil {

		switch {
		case errors.Is(err, apperrors.ErrInvalidID):
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Status: "error",
				Error:  "invalid file id",
			})

		case errors.Is(err, apperrors.ErrFileNotFound):
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Status: "error",
				Error:  "file not found",
			})

		default:
			h.Logger.Errorw(
				"failed to generate signed url",
				"file_id", fileID,
				"error", err,
			)

			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "failed to generate access url",
			})
		}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "signed url generated successfully (valid for limited time)",
		Data: map[string]string{
			"signed_url": signedURL,
		},
	})
}

func (h *FileHandler) CreatePrintJobsHandler(c echo.Context)error{
	ctx := c.Request().Context()

	h.Logger.Infow(
		"fetch print jobs request received",
	)

	var payload domain.PrintJobPayload

	if err := utils.DecodeAndValidateJSON(c.Request().Body, &payload); err != nil {
		h.Logger.Warnf("Invalid print payload | IP=%s | Error=%v", c.RealIP(), err)
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Status: "error",
			Error:  err.Error(),
		})
	}

	err := h.FileService.CreatePrintJobService(ctx, payload)
	if err != nil {
		
		switch {
		case errors.Is(err, apperrors.ErrInvalidID):
			h.Logger.Warnf("Invalid ObjectID formate | Error=%v", err)
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Status: "error",
				Error:  "Invalid FileID.",
			})
		case errors.Is(err, apperrors.ErrInvalidCopies):
			h.Logger.Warnf("Invalid copies | Error=%v", err)
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Status: "error",
				Error:  "Copies value must be between 1 and 100.",
			})
		case errors.Is(err, apperrors.ErrFileNotFound):
			h.Logger.Warnf("File not found in the Databse |Error=%v", err)
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Status: "error",
				Error:  "File not found.",
			})
		default:
			h.Logger.Errorf("Unexpected error while creating printJob | err=%v", err)
			return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Status: "error",
				Error:  "Internal error creating print job.",
			})
		}
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Print job created successfully",
		Data:    "print job created successfully",
	})
}

func (h *FileHandler) FetchPrintJobsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	h.Logger.Infow(
		"fetch print jobs request received",
	)

	printJobs, err := h.FileService.FetchPrintJobsService(ctx)
	if err != nil {
		h.Logger.Errorf("Failed to fetch print jobs | Error=%v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Failed to fetch print jobs",
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Print jobs fetched successfully",
		Data:    printJobs,
	})
}

func (h *FileHandler) CalculateTotalRevenueHandler(c echo.Context)error{
	ctx := c.Request().Context()
	
	h.Logger.Infow(
		"calculate total revenue request received",
	)

	totalRevenue, err := h.FileService.CalculateTotalRevenueService(ctx)
	if err != nil {
		h.Logger.Errorf("Failed to calculate total revenue | Error=%v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Failed to calculate total revenue",
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Total revenue calculated successfully",
		Data:    map[string]int{"total_revenue": totalRevenue},
	})
}

func (h *FileHandler) GetRecentPrintJobsHandler(c echo.Context) error {
	ctx := c.Request().Context()
	
	h.Logger.Infow(
		"fetch recent print jobs request received",
	)

	printJobs, err := h.FileService.GetRecentPrintJobsService(ctx)
	if err != nil {
		h.Logger.Errorf("Failed to fetch recent print jobs | Error=%v", err)
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Status: "error",
			Error:  "Failed to fetch recent print jobs",
		})
	}

	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Status:  "success",
		Message: "Recent print jobs fetched successfully",
		Data:    printJobs,
	})
}