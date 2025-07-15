package handlers

import (
        "bytes"
        "context"
        "fmt"
        "io"
        "log"
        "net/http"
        "time"
        "encoding/json"
        "github.com/gin-gonic/gin"
        "its-gateway/internal/config"
        "its-gateway/internal/dahua"
        "its-gateway/internal/mq"
)

type ITSHandler struct {
        Publisher *mq.Publisher
        Config    *config.Config
}

// ======================
// Dahua ITS JSON schema
// ======================

type PicData struct {
        PicName string `json:"PicName"`
        Content string `json:"Content"`
}

type Picture struct {
        NormalPic  *PicData `json:"NormalPic"`
        CutoutPic  *PicData `json:"CutoutPic"`
        VehiclePic *PicData `json:"VehiclePic"`
}

type Plate struct {
        PlateNumber string `json:"PlateNumber"`
}

type SnapInfo struct {
        AllowUser bool `json:"AllowUser"`
}

type DahuaEvent struct {
        Plate    Plate    `json:"Plate"`
        Picture  Picture  `json:"Picture"`
        SnapInfo SnapInfo `json:"SnapInfo"`
}

// ============================
// Handle POST from ITS camera
// ============================

func (h *ITSHandler) HandleITSEvent(c *gin.Context) {
        bodyBytes, err := io.ReadAll(c.Request.Body)
        if err != nil || len(bodyBytes) == 0 {
                log.Printf("[WARN] Invalid or empty request body: %v", err)
                c.JSON(http.StatusBadRequest, gin.H{"Result": false, "Message": "Invalid or empty body"})
                return
        }
        c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

        var event DahuaEvent
        if err := json.Unmarshal(bodyBytes, &event); err != nil {
                log.Printf("[WARN] Malformed Dahua ITS event: %v", err)
                c.JSON(http.StatusBadRequest, gin.H{"Result": false, "Message": "Malformed event"})
                return
        }

        log.Printf("[INFO] Received event | Plate: %s | AllowUser: %v",
                event.Plate.PlateNumber, event.SnapInfo.AllowUser)

        ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
        defer cancel()

        if err := h.Publisher.Publish(ctx, bodyBytes); err != nil {
                log.Printf("[ERROR] Failed to publish event for plate %s: %v", event.Plate.PlateNumber, err)
                c.JSON(http.StatusInternalServerError, gin.H{"Result": false, "Message": "Failed to queue event"})
                return
        }

        log.Printf("[INFO] ITS event for plate %s queued successfully", event.Plate.PlateNumber)
        c.JSON(http.StatusOK, gin.H{"Result": true, "Message": "Event queued"})
}

// ===================================
// Endpoint để mở barrier từ backend
// ===================================

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
