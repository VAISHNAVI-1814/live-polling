package controllers

import (
	"errors"
	"net/http"

	"live-polling-backend/models"
	"live-polling-backend/services"
	"live-polling-backend/websocket"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PollController struct {
	pollService *services.PollService
	wsHub       *websocket.Hub
}

func NewPollController(pollService *services.PollService, wsHub *websocket.Hub) *PollController {
	return &PollController{
		pollService: pollService,
		wsHub:       wsHub,
	}
}

func (ctrl *PollController) CreatePoll(c *gin.Context) {
	userId, _ := c.Get("userId")
	userEmail, _ := c.Get("userEmail")

	var req models.CreatePollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll data: " + err.Error()})
		return
	}

	creatorName := ""
	if emailStr, ok := userEmail.(string); ok {
		creatorName = emailStr
	}

	poll, err := ctrl.pollService.CreatePoll(c.Request.Context(), userId.(primitive.ObjectID), creatorName, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, poll)
}

func (ctrl *PollController) GetMyPolls(c *gin.Context) {
	userId, _ := c.Get("userId")
	polls, err := ctrl.pollService.GetUserPolls(c.Request.Context(), userId.(primitive.ObjectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch polls: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, polls)
}

func (ctrl *PollController) GetPoll(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	poll, err := ctrl.pollService.GetPollByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (ctrl *PollController) GetPollByShareCode(c *gin.Context) {
	shareCode := c.Param("shareCode")
	poll, err := ctrl.pollService.GetPollByShareCode(c.Request.Context(), shareCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found with share code: " + shareCode})
		return
	}

	c.JSON(http.StatusOK, poll)
}

func (ctrl *PollController) Vote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	var req models.VoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vote request: " + err.Error()})
		return
	}

	clientIP := c.ClientIP()
	results, err := ctrl.pollService.Vote(c.Request.Context(), id, req.OptionID, req.VoterToken, clientIP)
	if err != nil {
		if errors.Is(err, services.ErrDuplicateVote) {
			c.JSON(http.StatusConflict, gin.H{"error": "You have already cast your vote on this poll"})
			return
		}
		if errors.Is(err, services.ErrPollClosed) || errors.Is(err, services.ErrPollExpired) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, services.ErrInvalidOption) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Vote recorded successfully",
		"results": results,
	})
}

func (ctrl *PollController) GetResults(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	results, err := ctrl.pollService.GetResults(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Poll not found"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (ctrl *PollController) CheckVoted(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	voterToken := c.Query("voterToken")
	if voterToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "voterToken query parameter required"})
		return
	}

	hasVoted := ctrl.pollService.HasVoted(c.Request.Context(), id, voterToken)
	c.JSON(http.StatusOK, gin.H{"hasVoted": hasVoted})
}

func (ctrl *PollController) UpdateStatus(c *gin.Context) {
	userId, _ := c.Get("userId")
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	var req models.UpdatePollStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status: " + err.Error()})
		return
	}

	err = ctrl.pollService.UpdateStatus(c.Request.Context(), id, userId.(primitive.ObjectID), req.Status)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll status updated successfully", "status": req.Status})
}

func (ctrl *PollController) DeletePoll(c *gin.Context) {
	userId, _ := c.Get("userId")
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid poll ID"})
		return
	}

	err = ctrl.pollService.DeletePoll(c.Request.Context(), id, userId.(primitive.ObjectID))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Poll deleted successfully"})
}

func (ctrl *PollController) ServeWS(c *gin.Context) {
	pollID := c.Param("id")
	ctrl.wsHub.ServeWS(c.Writer, c.Request, pollID)
}
