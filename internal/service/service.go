package service

import (
	"fmt"

	"github.com/google/uuid"
	"kivo-player_sync-server/internal/models"
	"kivo-player_sync-server/internal/repository"
)

type TrackService struct {
	repo *repository.TrackRepository
}

func NewTrackService(repo *repository.TrackRepository) *TrackService {
	return &TrackService{repo: repo}
}

// SyncRequest — один трек от клиента.
type SyncRequest struct {
	Fingerprint string `json:"fingerprint" binding:"required"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Album       string `json:"album"`
	Duration    int    `json:"duration"`
}

// SyncResponse — ответ для одного трека.
type SyncResponse struct {
	Fingerprint     string `json:"fingerprint"`
	HasLyrics       bool   `json:"has_lyrics"`
	LyricsUpdatedAt string `json:"lyrics_updated_at,omitempty"`
}

// Sync принимает список треков от клиента, upsert'ит их и возвращает
// информацию о наличии текстов.
func (s *TrackService) Sync(items []SyncRequest) ([]SyncResponse, error) {
	fps := make([]string, 0, len(items))

	for _, item := range items {
		track := &models.Track{
			Fingerprint: item.Fingerprint,
			Title:       item.Title,
			Artist:      item.Artist,
			Album:       item.Album,
			Duration:    item.Duration,
		}
		if err := s.repo.UpsertTrack(track); err != nil {
			return nil, fmt.Errorf("upsert track %s: %w", item.Fingerprint, err)
		}
		fps = append(fps, item.Fingerprint)
	}

	withLyrics, err := s.repo.GetFingerprintsWithLyrics(fps)
	if err != nil {
		return nil, fmt.Errorf("get lyrics info: %w", err)
	}

	lyricsMap := make(map[string]string, len(withLyrics))
	for _, v := range withLyrics {
		lyricsMap[v.Fingerprint] = v.LyricsUpdatedAt
	}

	result := make([]SyncResponse, 0, len(items))
	for _, item := range items {
		updatedAt, has := lyricsMap[item.Fingerprint]
		result = append(result, SyncResponse{
			Fingerprint:     item.Fingerprint,
			HasLyrics:       has,
			LyricsUpdatedAt: updatedAt,
		})
	}

	return result, nil
}

// GetLyrics возвращает текст по fingerprint.
func (s *TrackService) GetLyrics(fingerprint string) (*models.Lyrics, error) {
	track, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil {
		return nil, err
	}
	if track == nil {
		return nil, nil
	}
	return s.repo.GetLyricsByTrackID(track.ID)
}

// UploadLyricsRequest — тело запроса на загрузку текста.
type UploadLyricsRequest struct {
	Content string `json:"content" binding:"required"`
	IsTimed bool   `json:"is_timed"`
	Source  string `json:"source"`
}

// UploadLyrics сохраняет текст для трека.
func (s *TrackService) UploadLyrics(fingerprint string, req UploadLyricsRequest) error {
	track, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil {
		return err
	}
	if track == nil {
		return fmt.Errorf("track not found")
	}

	source := req.Source
	if source == "" {
		source = "manual"
	}

	lyrics := &models.Lyrics{
		ID:      uuid.New(),
		TrackID: track.ID,
		Content: req.Content,
		IsTimed: req.IsTimed,
		Source:  source,
	}
	return s.repo.UpsertLyrics(lyrics)
}
