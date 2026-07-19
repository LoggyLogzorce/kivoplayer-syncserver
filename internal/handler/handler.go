package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"kivo-player_sync-server/internal/service"
)

type TrackHandler struct {
	svc *service.TrackService
}

func NewTrackHandler(svc *service.TrackService) *TrackHandler {
	return &TrackHandler{svc: svc}
}

// POST /api/sync
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

// GET /api/tracks/:fingerprint/lyrics
func (h *TrackHandler) GetLyrics(c *gin.Context) {
	fp := c.Param("fingerprint")

	lyrics, err := h.svc.GetLyrics(fp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if lyrics == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content":    lyrics.Content,
		"is_timed":   lyrics.IsTimed,
		"source":     lyrics.Source,
		"updated_at": lyrics.UpdatedAt,
	})
}

// POST /api/tracks/:fingerprint/lyrics
// Тело: {content, is_timed, source}
func (h *TrackHandler) UploadLyrics(c *gin.Context) {
	fp := c.Param("fingerprint")

	var req service.UploadLyricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UploadLyrics(fp, req); err != nil {
		if err.Error() == "track not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "track not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
