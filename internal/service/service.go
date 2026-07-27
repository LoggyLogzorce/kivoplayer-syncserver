package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"kivo-player_sync-server/internal/models"
	"kivo-player_sync-server/internal/repository"
)

var ErrFileAlreadyExists = errors.New("file already exists")

type TrackService struct {
	repo *repository.TrackRepository
}

func NewTrackService(repo *repository.TrackRepository) *TrackService {
	return &TrackService{repo: repo}
}

// SyncRequest — один трек от клиента.
type SyncRequest struct {
	AudioFingerPrint string `json:"audio_finger_print"`
	LrcFingerPrint   string `json:"lrc_finger_print"`
	Filename         string `json:"filename"`
}

// SyncResponse — ответ для одного трека.
type SyncResponse struct {
	Download []SyncFile `json:"download"`
	Upload   []SyncFile `json:"upload"`
}

type SyncFile struct {
	Type             string `json:"type"` // audio | lyrics
	AudioFingerprint string `json:"audio_fingerprint,omitempty"`
	LrcFingerprint   string `json:"lrc_fingerprint,omitempty"`
	FileName         string `json:"filename"`
}

// Sync принимает список треков от клиента, upsert'ит их и возвращает
// информацию о наличии текстов.
func (s *TrackService) Sync(req []SyncRequest) (*SyncResponse, error) {
	result := &SyncResponse{}

	// Получаем данные с сервера.
	tracks, err := s.repo.GetAllTracks()
	if err != nil {
		return nil, err
	}

	lyrics, err := s.repo.GetAllLyrics()
	if err != nil {
		return nil, err
	}

	// Карты клиента.
	clientAudio := make(map[string]string)  // fingerprint -> filename
	clientLyrics := make(map[string]string) // fingerprint -> filename

	for _, item := range req {
		if item.AudioFingerPrint != "" {
			clientAudio[item.AudioFingerPrint] = item.Filename
		}

		if item.LrcFingerPrint != "" {
			clientLyrics[item.LrcFingerPrint] = item.Filename
		}
	}

	// Карты сервера.
	serverAudio := make(map[string]models.Track)
	serverLyrics := make(map[string]models.Lyrics)

	for _, track := range tracks {
		serverAudio[track.Fingerprint] = track
	}

	for _, lyric := range lyrics {
		serverLyrics[lyric.Fingerprint] = lyric
	}

	// ---------- Upload ----------
	// Есть у клиента, нет на сервере.

	for fp, filename := range clientAudio {
		if _, ok := serverAudio[fp]; !ok {
			result.Upload = append(result.Upload, SyncFile{
				Type:             "audio",
				AudioFingerprint: fp,
				FileName:         filename,
			})
		}
	}

	for fp, filename := range clientLyrics {
		if _, ok := serverLyrics[fp]; !ok {
			result.Upload = append(result.Upload, SyncFile{
				Type:           "lyrics",
				LrcFingerprint: fp,
				FileName:       filename,
			})
		}
	}

	// ---------- Download ----------
	// Есть на сервере, нет у клиента.

	for fp, track := range serverAudio {
		if _, ok := clientAudio[fp]; !ok {
			result.Download = append(result.Download, SyncFile{
				Type:             "audio",
				AudioFingerprint: fp,
				FileName:         track.FileName,
			})
		}
	}

	for fp, lyric := range serverLyrics {
		if _, ok := clientLyrics[fp]; !ok {
			result.Download = append(result.Download, SyncFile{
				Type:           "lyrics",
				LrcFingerprint: fp,
				FileName:       lyric.FileName,
			})
		}
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

// UploadLyrics сохраняет LRC файл для трека и запись в БД.
func (s *TrackService) UploadLyrics(audioFp, lrcFp, filename string, content []byte) error {
	track, err := s.repo.GetByFingerprint(audioFp)
	if err != nil {
		return err
	}
	if track == nil {
		return fmt.Errorf("track not found")
	}
	if track.FileName == "" {
		return fmt.Errorf("track file name not set")
	}

	path := "internal/storage/lyrics/" + filename

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrFileAlreadyExists, filename)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	lyrics := &models.Lyrics{
		ID:          uuid.New(),
		TrackID:     track.ID,
		Fingerprint: lrcFp,
		FileName:    filename,
	}
	return s.repo.UpsertLyrics(lyrics)
}

// GetLyricsPath возвращает путь к LRC файлу для трека.
func (s *TrackService) GetLyricsPath(fingerprint string) (string, error) {
	track, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil {
		return "", err
	}
	if track == nil {
		return "", fmt.Errorf("track not found")
	}

	lyrics, err := s.repo.GetLyricsByTrackID(track.ID)
	if err != nil {
		return "", err
	}
	if lyrics == nil {
		return "", fmt.Errorf("lyrics not found")
	}

	return "internal/storage/lyrics/" + lyricsFilename(track.FileName), nil
}

// UploadTrack сохраняет аудио файл для трека и обновляет FileName.
func (s *TrackService) UploadTrack(fingerprint string, content []byte, fileName string) error {
	track, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil {
		return err
	}

	path := "internal/storage/music/" + fileName
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%w: %s", ErrFileAlreadyExists, fileName)
	}

	if track == nil {
		track = &models.Track{ID: uuid.New(), Fingerprint: fingerprint}
		if err := s.repo.UpsertTrack(track); err != nil {
			return fmt.Errorf("create track: %w", err)
		}
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("save file: %w", err)
	}

	return s.repo.SetTrackFileName(track.ID, fileName)
}

// GetTrackPath возвращает путь к аудио файлу трека.
func (s *TrackService) GetTrackPath(fingerprint string) (string, error) {
	track, err := s.repo.GetByFingerprint(fingerprint)
	if err != nil {
		return "", err
	}
	if track == nil {
		return "", fmt.Errorf("track not found")
	}
	if track.FileName == "" {
		return "", fmt.Errorf("track file name not set")
	}
	return filepath.Join("internal", "storage", "music", track.FileName), nil
}

func lyricsFilename(audioName string) string {
	return audioName[:len(audioName)-len(filepath.Ext(audioName))] + ".lrc"
}
