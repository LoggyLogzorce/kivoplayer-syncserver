package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"kivo-player_sync-server/internal/service"
)

type TrackHandler struct {
	svc *service.TrackService
}

func NewTrackHandler(svc *service.TrackService) *TrackHandler {
	return &TrackHandler{svc: svc}
}

// Sync POST /api/sync
// Тело: [{fingerprint, title, artist, album, duration}, ...]
func (h *TrackHandler) Sync(c *gin.Context) {
	var req []service.SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty list"})
		return
	}

	result, err := h.svc.Sync(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// UploadTrack POST /api/tracks/:audio-fingerprint
// Форма: file=song.mp3
func (h *TrackHandler) UploadTrack(c *gin.Context) {
	fp := c.Param("audio-fingerprint")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file"})
		return
	}

	if err := h.svc.UploadTrack(fp, content, header.Filename); err != nil {
		if errors.Is(err, service.ErrFileAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetTrack GET /api/tracks/:audio-fingerprint
func (h *TrackHandler) GetTrack(c *gin.Context) {
	fp := c.Param("audio-fingerprint")

	filePath, err := h.svc.GetTrackPath(fp)
	if err != nil {
		if err.Error() == "track not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.FileAttachment(filePath, filepath.Base(filePath))
}

// UploadLyrics POST /api/tracks/:audio-fingerprint/lyrics/:lrc-fingerprint
// Форма: file=song.lrc
func (h *TrackHandler) UploadLyrics(c *gin.Context) {
	audioFp := c.Param("audio-fingerprint")
	lrcFp := c.Param("lrc-fingerprint")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "read file"})
		return
	}

	if err := h.svc.UploadLyrics(audioFp, lrcFp, header.Filename, content); err != nil {
		if errors.Is(err, service.ErrFileAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "track not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// GetLyrics GET /api/tracks/:audio-fingerprint/lyrics/:lrc-fingerprint
func (h *TrackHandler) GetLyrics(c *gin.Context) {
	fp := c.Param("audio-fingerprint")

	filePath, err := h.svc.GetLyricsPath(fp)
	if err != nil {
		if err.Error() == "track not found" || err.Error() == "lyrics not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.FileAttachment(filePath, filepath.Base(filePath))
}
