package plant

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ScanService struct {
	upgrader websocket.Upgrader
}

func NewScanService() *ScanService {
	return &ScanService{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow connections from any origin (configure properly for production)
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

// Message types for WebSocket communication
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type FrameData struct {
	Image     string `json:"image"`     // base64 encoded image
	Timestamp int64  `json:"timestamp"` // Unix timestamp
}

type PredictionResult struct {
	Prediction string  `json:"prediction"`
	Confidence float64 `json:"confidence"`
	ProcessedAt int64  `json:"processed_at"`
}

// ScanImageHandler godoc
// @Summary      Process scanned image
// @Description  Accepts real-time image uploads for scanning
// @Tags         Scanner
// @Accept       mpfd
// @Produce      json
// @Param        file formData file true "Image to scan"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]string
// @Failure      500 {object} map[string]string
// @Router       /scan/image [post]
func (s *ScanService) ScanImageHandler(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Optionally save to disk or process directly (simulate processing)
	fmt.Printf("Received file: %s at %s\n", file.Filename, time.Now().Format(time.RFC3339))

	// In a real system, you would:
	// - Open the file
	// - Decode image
	// - Run through prediction model
	// - Return result

	c.JSON(http.StatusOK, gin.H{
		"status":   "processed",
		"filename": file.Filename,
		// Simulated result
		"prediction": "Plant: Ficus lyrata",
	})
}

// ScanVideoHandler godoc
// @Summary      Process live video stream
// @Description  WebSocket endpoint for real-time video frame processing
// @Tags         Scanner
// @Accept       json
// @Produce      json
// @Success      101 {string} string "Switching Protocols"
// @Router       /scan/video [get]
func (s *ScanService) ScanVideoHandler(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connection established from %s", conn.RemoteAddr())

	// Send initial connection confirmation
	welcomeMsg := WSMessage{
		Type: "connected",
		Data: map[string]interface{}{
			"message":    "Connected to plant scanner",
			"timestamp":  time.Now().Unix(),
			"session_id": generateSessionID(),
		},
	}
	
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return
	}

	// Handle incoming messages
	for {
		var msg WSMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		switch msg.Type {
		case "frame":
			s.handleFrame(conn, msg)
		case "ping":
			s.handlePing(conn)
		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

func (s *ScanService) handleFrame(conn *websocket.Conn, msg WSMessage) {
	// Parse frame data
	frameDataMap, ok := msg.Data.(map[string]interface{})
	if !ok {
		s.sendError(conn, "Invalid frame data format")
		return
	}

	imageData, ok := frameDataMap["image"].(string)
	if !ok {
		s.sendError(conn, "Missing or invalid image data")
		return
	}

	timestamp, ok := frameDataMap["timestamp"].(float64)
	if !ok {
		timestamp = float64(time.Now().Unix())
	}

	// Validate base64 image data
	if _, err := base64.StdEncoding.DecodeString(imageData); err != nil {
		s.sendError(conn, "Invalid base64 image data")
		return
	}

	// Process the frame (simulate plant detection)
	result := s.processFrame(imageData, int64(timestamp))

	// Send result back to client
	response := WSMessage{
		Type: "prediction",
		Data: result,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending prediction: %v", err)
	}
}

func (s *ScanService) handlePing(conn *websocket.Conn) {
	pong := WSMessage{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}
	
	if err := conn.WriteJSON(pong); err != nil {
		log.Printf("Error sending pong: %v", err)
	}
}

func (s *ScanService) sendError(conn *websocket.Conn, message string) {
	errorMsg := WSMessage{
		Type: "error",
		Data: map[string]interface{}{
			"message":   message,
			"timestamp": time.Now().Unix(),
		},
	}
	
	if err := conn.WriteJSON(errorMsg); err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

func (s *ScanService) processFrame(imageData string, timestamp int64) PredictionResult {
	// Simulate processing time
	time.Sleep(50 * time.Millisecond)

	// In a real implementation, you would:
	// 1. Decode the base64 image
	// 2. Preprocess the image (resize, normalize, etc.)
	// 3. Run inference through your ML model
	// 4. Return the prediction with confidence score

	// Simulated predictions for demo
	predictions := []string{
		"Ficus lyrata (Fiddle Leaf Fig)",
		"Monstera deliciosa",
		"Pothos aureus",
		"Sansevieria trifasciata",
		"Philodendron hederaceum",
	}

	// Simulate varying confidence
	confidence := 0.7 + (float64(time.Now().UnixNano()%30) / 100.0)
	predictionIndex := int(timestamp) % len(predictions)

	return PredictionResult{
		Prediction:  predictions[predictionIndex],
		Confidence:  confidence,
		ProcessedAt: time.Now().Unix(),
	}
}

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}