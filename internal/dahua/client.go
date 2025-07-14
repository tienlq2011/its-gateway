package dahua

import (
	"context"
	"fmt"
	"net/http"
	"time"

	digest "github.com/xinsnake/go-http-digest-auth-client"
)

// Client represents a Dahua device controller
type Client struct {
	Username string
	Password string
	LaneMap  map[string]string
	Timeout  time.Duration
}

// NewClient creates a Dahua client with lane-to-IP map
func NewClient(username, password string, laneMap map[string]string) *Client {
	return &Client{
		Username: username,
		Password: password,
		LaneMap:  laneMap,
		Timeout:  5 * time.Second,
	}
}

// OpenDoor triggers a Dahua device to open gate for a given lane ID
func (c *Client) OpenDoor(ctx context.Context, laneID string) error {
	ip, ok := c.LaneMap[laneID]
	if !ok || ip == "" {
		return fmt.Errorf("no Dahua IP configured for lane: %s", laneID)
	}

	url := fmt.Sprintf("http://%s/cgi-bin/accessControl.cgi?action=openDoor&channel=1&UserID=101", ip)

	// Create digest-auth request
	req := digest.NewRequest(c.Username, c.Password, "GET", url, "")

	// Execute the request (library uses its own internal http.Client)
	resp, err := req.Execute()
	if err != nil {
		return fmt.Errorf("failed to send openDoor command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Dahua responded with unexpected status: %d", resp.StatusCode)
	}

	return nil
}

