package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"its-gateway/internal/config"
	"its-gateway/internal/dahua"
	"its-gateway/internal/mq"
)

type ITSHandler struct {
	Publisher *mq.Publisher
	Config    *config.Config
}

type ITSEvent struct {
	EventType string `json:"event_type"`
	CameraID  string `json:"camera_id"`
	Timestamp string `json:"timestamp"`
	// Add more fields if needed
}

// HandleITSEvent receives ITS camera events and pushes them to RabbitMQ
func (h *ITSHandler) HandleITSEvent(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bodyBytes) == 0 {
		log.Printf("[WARN] Invalid or empty request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"Result": false, "Message": "Invalid or empty body"})
		return
	}
	// Reset request body so it can be reused if needed
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var event ITSEvent
	if err := json.Unmarshal(bodyBytes, &event); err != nil {
		log.Printf("[WARN] Malformed ITS event payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"Result": false, "Message": "Malformed ITS event"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := h.Publisher.Publish(ctx, bodyBytes); err != nil {
		log.Printf("[ERROR] Failed to publish ITS event from camera %s: %v", event.CameraID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"Result": false, "Message": "Failed to queue event"})
		return
	}

	log.Printf("[INFO] ITS event from camera %s queued successfully", event.CameraID)
	c.JSON(http.StatusOK, gin.H{"Result": true, "Message": "Event queued"})
}

// OpenBarrier sends a command to Dahua camera to open the gate at the specified lane
func (h *ITSHandler) OpenBarrier(c *gin.Context) {
	laneID := c.Param("lane")
	if laneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"Result": false, "Message": "Missing lane ID"})
		return
	}

	dahuaClient := dahua.NewClient(
		h.Config.Dahua.Username,
		h.Config.Dahua.Password,
		h.Config.Dahua.LaneMap,
	)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := dahuaClient.OpenDoor(ctx, laneID); err != nil {
		log.Printf("[ERROR] Failed to open barrier for lane '%s': %v", laneID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"Result": false, "Message": "Failed to open gate"})
		return
	}

	log.Printf("[INFO] Barrier opened successfully for lane '%s'", laneID)
	c.JSON(http.StatusOK, gin.H{"Result": true, "Message": fmt.Sprintf("Barrier opened for lane %s", laneID)})
}

