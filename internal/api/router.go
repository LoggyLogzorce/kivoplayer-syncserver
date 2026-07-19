package api

import (
	"github.com/gin-gonic/gin"
	"kivo-player_sync-server/internal/handler"
)

func New(trackHandler *handler.TrackHandler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/sync", trackHandler.Sync)

		tracks := api.Group("/tracks")
		{
			tracks.GET("/:fingerprint/lyrics", trackHandler.GetLyrics)
			tracks.POST("/:fingerprint/lyrics", trackHandler.UploadLyrics)
		}
	}

	return r
}
