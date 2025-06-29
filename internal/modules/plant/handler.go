package plant

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"bytes"

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

// WSMessage is for WebSocket communication
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// PredictionResult holds the model prediction
type PredictionResult struct {
	Prediction  string  `json:"prediction"`
	Confidence  float64 `json:"confidence"`
	ProcessedAt int64   `json:"processed_at"`
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

	openedFile, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer openedFile.Close()

	fileBytes := make([]byte, file.Size)
	_, err = openedFile.Read(fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read uploaded file"})
		return
	}

	base64Str := base64.StdEncoding.EncodeToString(fileBytes)
	result := s.processFrame(base64Str)

	c.JSON(http.StatusOK, gin.H{
		"status":     "processed",
		"filename":   file.Filename,
		"prediction": result.Prediction,
		"confidence": result.Confidence,
		"processed":  result.ProcessedAt,
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

	if _, err := base64.StdEncoding.DecodeString(imageData); err != nil {
		s.sendError(conn, "Invalid base64 image data")
		return
	}

	result := s.processFrame(imageData)

	response := WSMessage{
		Type: "prediction",
		Data: result,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending prediction: %v", err)
	}
}


func (s *ScanService) processFrame(imageData string) PredictionResult {
    tmpFile, err := os.CreateTemp("", "plant_image_*.txt")
    if err != nil {
        log.Printf("Failed to create temp file: %v", err)
        return PredictionResult{
            Prediction:  "Internal Error",
            Confidence:  0.0,
            ProcessedAt: time.Now().Unix(),
        }
    }
    defer os.Remove(tmpFile.Name())

    _, err = tmpFile.WriteString(imageData)
    if err != nil {
        log.Printf("Failed to write temp file: %v", err)
        return PredictionResult{
            Prediction:  "Internal Error",
            Confidence:  0.0,
            ProcessedAt: time.Now().Unix(),
        }
    }
    tmpFile.Close()

    cmd := exec.Command("python3", "ml/predict.py", tmpFile.Name())

    var stderr bytes.Buffer
    cmd.Stderr = &stderr

    output, err := cmd.Output()
    if err != nil {
        log.Printf("Prediction failed: %v, stderr: %s", err, stderr.String())
        return PredictionResult{
            Prediction:  "Prediction Error",
            Confidence:  0.0,
            ProcessedAt: time.Now().Unix(),
        }
    }

    result := strings.TrimSpace(string(output))
    parts := strings.Split(result, "|")
    if len(parts) != 2 {
        log.Printf("Unexpected prediction output: %s", result)
        return PredictionResult{
            Prediction:  "Unexpected Output",
            Confidence:  0.0,
            ProcessedAt: time.Now().Unix(),
        }
    }

    confidence, err := strconv.ParseFloat(parts[1], 64)
    if err != nil {
        log.Printf("Invalid confidence value: %s", parts[1])
        confidence = 0.0
    }

    return PredictionResult{
        Prediction:  parts[0],
        Confidence:  confidence,
        ProcessedAt: time.Now().Unix(),
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

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}
