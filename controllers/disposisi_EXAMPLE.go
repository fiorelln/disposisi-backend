package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/fiorelln/disposisi/dto"
	"github.com/fiorelln/disposisi/services"
	"github.com/gin-gonic/gin"
)

type DisposisiController struct {
	service services.DisposisiService
}

func NewDisposisiController(service services.DisposisiService) *DisposisiController {
	return &DisposisiController{
		service: service,
	}
}


func (c *DisposisiController) ForwardDisposisi(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	disposisiID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid disposisi ID",
		})
		return
	}

	var req dto.CreateForwardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	canForward, reason, err := c.service.ValidateForward(
		uint(disposisiID),
		req.ToUserID,
		currentUserID.(uint),
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: fmt.Sprintf("Validation failed: %v", err),
		})
		return
	}

	if !canForward {
		ctx.JSON(http.StatusForbidden, dto.ErrorResponse{
			Code:    "FORBIDDEN",
			Message: reason,
		})
		return
	}

	newDisposisi, err := c.service.ForwardDisposisi(
		uint(disposisiID),
		&req,
		currentUserID.(uint),
	)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "FORWARD_FAILED",
			Message: fmt.Sprintf("Failed to forward: %v", err),
		})
		return
	}

	response, _ := c.service.GetDisposisiDetail(newDisposisi.ID)
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Disposisi forwarded successfully",
		"data":    response,
	})
}


func (c *DisposisiController) CompleteDisposisi(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	disposisiID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid disposisi ID",
		})
		return
	}

	var req dto.CompleteDisposisiRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
		})
		return
	}

	err = c.service.CompleteDisposisi(
		uint(disposisiID),
		&req,
		currentUserID.(uint),
	)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "COMPLETE_FAILED",
			Message: fmt.Sprintf("Failed to complete: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Disposisi completed successfully",
	})
}

func (c *DisposisiController) RejectDisposisi(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	disposisiID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid disposisi ID",
		})
		return
	}

	type RejectRequest struct {
		Reason string `json:"reason" binding:"required"`
	}

	var req RejectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Reason is required",
		})
		return
	}

	err = c.service.RejectDisposisi(
		uint(disposisiID),
		req.Reason,
		currentUserID.(uint),
	)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "REJECT_FAILED",
			Message: fmt.Sprintf("Failed to reject: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Disposisi rejected successfully",
	})
}


func (c *DisposisiController) MarkAsRead(ctx *gin.Context) {
	disposisiID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid disposisi ID",
		})
		return
	}

	err = c.service.MarkAsRead(uint(disposisiID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "MARK_FAILED",
			Message: fmt.Sprintf("Failed to mark as read: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Disposisi marked as read",
	})
}


func (c *DisposisiController) MarkAsReadBatch(ctx *gin.Context) {
	var req struct {
		DisposisiIDs []uint `json:"disposisi_ids" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "disposisi_ids is required",
		})
		return
	}

	err := c.service.MarkAsReadBatch(req.DisposisiIDs)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "MARK_FAILED",
			Message: fmt.Sprintf("Failed to mark as read: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Marked %d disposisi as read", len(req.DisposisiIDs)),
	})
}

func (c *DisposisiController) GetInbox(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	page := ctx.DefaultQuery("page", "1")
	pageSize := ctx.DefaultQuery("page_size", "20")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt < 1 || pageSizeInt > 100 {
		pageSizeInt = 20
	}

	response, err := c.service.GetInbox(currentUserID.(uint), pageInt, pageSizeInt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INBOX_ERROR",
			Message: fmt.Sprintf("Failed to get inbox: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}


func (c *DisposisiController) GetSentItems(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	page := ctx.DefaultQuery("page", "1")
	pageSize := ctx.DefaultQuery("page_size", "20")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeInt < 1 || pageSizeInt > 100 {
		pageSizeInt = 20
	}

	response, err := c.service.GetSentItems(currentUserID.(uint), pageInt, pageSizeInt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "SENT_ERROR",
			Message: fmt.Sprintf("Failed to get sent items: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}


func (c *DisposisiController) GetHistory(ctx *gin.Context) {
	suratID, err := strconv.ParseUint(ctx.Param("surat_id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid surat ID",
		})
		return
	}

	response, err := c.service.GetHistory(uint(suratID))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "HISTORY_ERROR",
			Message: fmt.Sprintf("Failed to get history: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *DisposisiController) GetDisposisiDetail(ctx *gin.Context) {
	disposisiID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_ID",
			Message: "Invalid disposisi ID",
		})
		return
	}

	response, err := c.service.GetDisposisiDetail(uint(disposisiID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Disposisi not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, response)
}


func (c *DisposisiController) GetStats(ctx *gin.Context) {
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	stats, err := c.service.GetStats(currentUserID.(uint))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "STATS_ERROR",
			Message: fmt.Sprintf("Failed to get stats: %v", err),
		})
		return
	}

	ctx.JSON(http.StatusOK, stats)
}
