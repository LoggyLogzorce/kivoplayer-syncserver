package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"kivo-player_sync-server/internal/models"
)

type TrackRepository struct {
	db *gorm.DB
}

func NewTrackRepository(db *gorm.DB) *TrackRepository {
	return &TrackRepository{db: db}
}

// UpsertTrack создаёт трек если не существует, иначе обновляет метаданные.
func (r *TrackRepository) UpsertTrack(track *models.Track) error {
	return r.db.Where(models.Track{Fingerprint: track.Fingerprint}).
		Assign(models.Track{
			Duration: track.Duration,
		}).
		FirstOrCreate(track).Error
}

// GetByFingerprint возвращает трек по fingerprint.
func (r *TrackRepository) GetByFingerprint(fp string) (*models.Track, error) {
	var track *models.Track
	err := r.db.Where("fingerprint = ?", fp).First(&track).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return track, err
}

// GetFingerprintsWithLyrics принимает список fingerprint'ов и возвращает те,
// для которых на сервере есть текст, с временем последнего обновления.
func (r *TrackRepository) GetFingerprintsWithLyrics(fps []string) ([]FingerprintLyricsInfo, error) {
	var result []FingerprintLyricsInfo

	err := r.db.Table("tracks").
		Select("tracks.fingerprint, lyrics.updated_at as lyrics_updated_at").
		Joins("JOIN lyrics ON lyrics.track_id = tracks.id").
		Where("tracks.fingerprint IN ?", fps).
		Scan(&result).Error

	return result, err
}

type FingerprintLyricsInfo struct {
	Fingerprint     string `json:"fingerprint"`
	LyricsUpdatedAt string `json:"lyrics_updated_at"`
}

// SetTrackFileName обновляет имя файла трека.
func (r *TrackRepository) SetTrackFileName(trackID uuid.UUID, fileName string) error {
	return r.db.Model(&models.Track{}).Where("id = ?", trackID).Update("file_name", fileName).Error
}

// GetLyricsByTrackID возвращает текст для трека.
func (r *TrackRepository) GetLyricsByTrackID(trackID uuid.UUID) (*models.Lyrics, error) {
	var lyrics models.Lyrics
	err := r.db.Where("track_id = ?", trackID).First(&lyrics).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &lyrics, err
}

// UpsertLyrics создаёт или перезаписывает текст для трека.
func (r *TrackRepository) UpsertLyrics(lyrics *models.Lyrics) error {
	var existing models.Lyrics
	err := r.db.Where("track_id = ?", lyrics.TrackID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(lyrics).Error
	}
	return err
}

func (r *TrackRepository) GetAllTracks() ([]models.Track, error) {
	var tracks []models.Track
	err := r.db.Find(&tracks).Error
	return tracks, err
}

func (r *TrackRepository) GetAllLyrics() ([]models.Lyrics, error) {
	var lyrics []models.Lyrics
	err := r.db.Find(&lyrics).Error
	return lyrics, err
}
