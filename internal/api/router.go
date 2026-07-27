package api

import (
	"github.com/gin-gonic/gin"
	"kivo-player_sync-server/internal/handler"
)

func New(trackHandler *handler.TrackHandler) *gin.Engine {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Next()
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
		return
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{})
	})

	r.RedirectTrailingSlash = false

	api := r.Group("/api")
	{
		api.POST("/sync/check", trackHandler.Sync)

		tracks := api.Group("/tracks/:audio-fingerprint")
		{
			tracks.POST("", trackHandler.UploadTrack)
			tracks.GET("", trackHandler.GetTrack)

			tracks.POST("/lyrics/:lrc-fingerprint", trackHandler.UploadLyrics)
			tracks.GET("/lyrics/:lrc-fingerprint", trackHandler.GetLyrics)
		}
	}

	return r
}
