package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"scriberr-backend/internal/database"
	"scriberr-backend/internal/models"

	"github.com/google/uuid"
)

// URLAudioDownloadRequest defines the structure for the URL audio download request.
type URLAudioDownloadRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// DownloadURLAudio handles the request to download and process audio from a URL.
func DownloadURLAudio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req URLAudioDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		writeJSONError(w, "Missing URL", http.StatusBadRequest)
		return
	}

	// Validate URL
	if !isValidURL(req.URL) {
		writeJSONError(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// Generate unique ID for this audio record
	recordID := uuid.NewString()

	// Set default title if not provided
	if req.Title == "" {
		req.Title = "Audio from URL"
	}

	// Ensure storage directories exist
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Printf("Error creating uploads directory: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if err := os.MkdirAll(convertedDir, 0755); err != nil {
		log.Printf("Error creating converted directory: %v", err)
		writeJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Define file paths
	extension := getFileExtension(req.URL)
	originalFileName := recordID + extension
	originalFilePath := filepath.Join(uploadsDir, originalFileName)
	convertedFileName := recordID + ".wav"
	convertedFilePath := filepath.Join(convertedDir, convertedFileName)

	// Step 1: Download audio from URL
	log.Printf("Starting download from URL: %s", req.URL)
	
	// Download the file
	err := downloadFile(originalFilePath, req.URL)
	if err != nil {
		log.Printf("Download failed for URL %s. Error: %v", req.URL, err)
		writeJSONError(w, fmt.Sprintf("Failed to download file: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully downloaded audio from URL")

	// Step 2: Insert metadata into database
	db := database.GetDB()
	audioRecord := models.Audio{
		ID:         recordID,
		Title:      req.Title,
		CreatedAt:  time.Now().UTC(),
		Transcript: "{}", // Default empty JSON object
		SpeakerMap: "{}", // Default empty JSON object
		Summary:    "{}", // Default empty JSON object
	}

	query := `INSERT INTO audio_records (id, title, transcript, speaker_map, summary, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	stmt, err := db.Prepare(query)
	if err != nil {
		log.Printf("Error preparing database statement: %v", err)
		// Cleanup downloaded file if DB preparation fails
		os.Remove(originalFilePath)
		writeJSONError(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	if _, err := stmt.Exec(audioRecord.ID, audioRecord.Title, audioRecord.Transcript, audioRecord.SpeakerMap, audioRecord.Summary, audioRecord.CreatedAt); err != nil {
		log.Printf("Error executing database insert: %v", err)
		// Cleanup downloaded file if DB insert fails
		os.Remove(originalFilePath)
		writeJSONError(w, "Failed to create record in database", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully inserted record %s into database", recordID)

	// Step 3: Convert audio to WAV using ffmpeg
	log.Printf("Converting downloaded audio to WAV format")
	// Command: ffmpeg -i "${inputPath}" -ar 16000 -ac 1 -c:a pcm_s16le "${outputPath}"
	convertCmd := exec.Command("ffmpeg", "-i", originalFilePath, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", convertedFilePath)

	convertOutput, err := convertCmd.CombinedOutput()
	if err != nil {
		log.Printf("ffmpeg conversion failed for %s. Error: %v. Output: %s", recordID, err, string(convertOutput))
		// If conversion fails, clean up the database record and downloaded file
		deleteQuery := "DELETE FROM audio_records WHERE id = ?"
		if _, delErr := db.Exec(deleteQuery, recordID); delErr != nil {
			log.Printf("CRITICAL: Failed to delete database record %s after ffmpeg failure: %v", recordID, delErr)
		}
		os.Remove(originalFilePath)
		writeJSONError(w, fmt.Sprintf("Failed to convert audio file: %s", string(convertOutput)), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully converted file to %s", convertedFilePath)

	// Step 4: Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": recordID})
}

// isValidURL checks if the provided string is a valid URL.
func isValidURL(urlStr string) bool {
	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return false
	}

	u, err := url.Parse(urlStr)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}

	// Only allow http and https schemes
	return u.Scheme == "http" || u.Scheme == "https"
}

// downloadFile downloads a file from a URL and saves it to a local path.
func downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// getFileExtension extracts the file extension from a URL.
func getFileExtension(urlStr string) string {
	// Parse the URL to handle query parameters and fragments
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ".mp3" // Default to .mp3 if URL parsing fails
	}

	// Get the path and split by '/' to get the last part
	pathParts := strings.Split(parsedURL.Path, "/")
	if len(pathParts) == 0 {
		return ".mp3"
	}

	// Get the last part which should be the filename
	filename := pathParts[len(pathParts)-1]
	
	// Split by '?' to remove query parameters if any
	filename = strings.Split(filename, "?")[0]
	
	// Get the file extension
	ext := filepath.Ext(filename)
	if ext == "" {
		return ".mp3" // Default to .mp3 if no extension found
	}
	
	return ext
}
