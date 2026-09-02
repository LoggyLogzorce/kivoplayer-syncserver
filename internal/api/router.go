package api

import (
	"github.com/gin-gonic/gin"
	"kivo-player_sync-server/internal/handler"
	"net/http"
)

func New(trackHandler *handler.TrackHandler, authKey string) *gin.Engine {
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
	api.Use(func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token != authKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "Invalid token"})
			return
		}
		c.Next()
	})
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
