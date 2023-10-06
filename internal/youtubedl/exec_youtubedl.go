package youtubedl

import (
	"errors"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/kkdai/youtube/v2"
	"github.com/minio/minio-go/v7"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

var playlistRegex = regexp.MustCompile(`^https://(www\.?)(youtube\..+)/playlist\?list=[a-zA-Z0-9_-]+$`)
var videoRegex = regexp.MustCompile(`^https://(www\.?)(youtube.+)/watch\?v=[a-zA-Z0-9_-]+`)

type ExecYouTubeDL struct {
	Configuration Configuration
	S3            *minio.Client
	MusicStore    *musicstore.MusicStore
	Log           *zap.Logger
	PlaylistStore *playliststore.PlaylistStore
}

func (e *ExecYouTubeDL) Save(p YouTubeDLParams) error {
	if isPlaylist(p) {
		return e.downloadPlaylist(p)
	} else if isVideo(p) {
		return e.downloadVideo(p)
	} else {
		err := errors.New("Cannot detect URL type")
		e.Log.Error(err.Error(), p.ZapFields()...)
		return err

	}
}

func isPlaylist(p YouTubeDLParams) bool {
	return playlistRegex.MatchString(p.Source)
}

func isVideo(p YouTubeDLParams) bool {
	return videoRegex.MatchString(p.Source)
}

func (e *ExecYouTubeDL) downloadVideo(p YouTubeDLParams) error {
	return e.download(p.Source, p.GuildID, p.PlaylistName)
}

func (e *ExecYouTubeDL) downloadPlaylist(p YouTubeDLParams) error {
	var err error

	ytclient := youtube.Client{}
	playlist, err := ytclient.GetPlaylist(p.Source)
	if err != nil {
		return err
	}

	for i := range playlist.Videos {
		newErr := e.download(
			idToUrl(playlist.Videos[i].ID),
			p.GuildID,
			p.PlaylistName,
		)
		err = multierr.Append(err, newErr)
	}

	return err
}

const TMPFILE_MP3 = "/tmp/audio.mp3"
const TMPFILE_M4A = "/tmp/audio.m4a"
const TMPFILE_DCA = "/tmp/audio.dca"
const TMPFILE_ART = "/tmp/art.jpg"

func (e *ExecYouTubeDL) download(source string, guildId string, playlistName string) error {
	log := e.Log.With(
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("playlistName", playlistName),
	)

	id, err := uuid.NewRandom()
	if err != nil {
		log.Warn("cannot create uuid")
		return err
	}

	uuid := id.String()
	log = log.With(zap.String("uuid", id.String()))

	log.Info("downloading")

	ytdl := ytdlCommand(source)
	ffmpegmp3 := ffmpegmp3Command()

	log.Info("downloading youtube file")

	output, err := ytdl.Output()
	if err != nil {
		log.Error("cannot fetch youtube file", zap.Error(err), zap.String("output", string(output)))
		return err
	}
	log.Info("convert to mp3")

	output, err = ffmpegmp3.Output()
	if err != nil {
		log.Error("cannot convert to mp3", zap.Error(err), zap.String("output", string(output)))
		return err
	}
	log.Info("convert to dca")

	err = convertToDca()
	if err != nil {
		log.Error("cannot convert to dca", zap.Error(err))
		return err
	}
	log.Info("fetch title")

	title, err := e.GetTitle(source)
	if err != nil {
		log.Error("cannot fetch title", zap.Error(err))
		return err
	}

	log.Info("fetch author")
	author, err := e.GetAuthor(source)
	if err != nil {
		log.Error("cannot fetch author", zap.Error(err))
		return err
	}
	log.Info("fetch thumbnail url")

	thumbnail, err := e.GetThumbnail(source)
	if err != nil {
		log.Error("cannot fetch AlbumArt", zap.Error(err))
		return err
	}
	t := &musicstore.Track{
		AlbumArt: thumbnail,
		Name:     title,
		Artist:   author,
		Source:   source,
		GuildID:  guildId,
		Uuid:     uuid,
	}
	log.Info("downloading thumbnail file")

	err = fetchThumbnail(thumbnail)
	if err != nil {
		log.Error("Failed to download AlbumArt", zap.Error(err))
		return err
	}
	log.Info("uploading mp3 to bucket")

	err = e.MusicStore.UploadMP3(t, TMPFILE_MP3)
	if err != nil {
		log.Error("Failed to upload mp3", zap.Error(err))
		return err
	}
	log.Info("uploading m4a to bucket")

	err = e.MusicStore.UploadM4A(t, TMPFILE_M4A)
	if err != nil {
		log.Error("Failed to upload m4a", zap.Error(err))
		return err
	}
	log.Info("uploading dca to bucket")

	err = e.MusicStore.UploadDCA(t, TMPFILE_DCA)
	if err != nil {
		log.Error("Failed to upload dca", zap.Error(err))
		return err
	}

	log.Info("uploading albumart to bucket")

	err = e.MusicStore.UploadAlbumArt(t, TMPFILE_ART)
	if err != nil {
		log.Error("Failed to upload albumart", zap.Error(err))
	}

	log.Info("uploading manifest to bucket")

	err = e.MusicStore.Save(t)
	if err != nil {
		log.Error("Failed to save track", zap.Error(err))
		return err
	}

	log.Info("done")

	err = multierr.Combine(
		os.Remove(TMPFILE_M4A),
		os.Remove(TMPFILE_MP3),
		os.Remove(TMPFILE_DCA),
		os.Remove(TMPFILE_ART),
	)
	if err != nil {
		return err
	}

	return e.PlaylistStore.Append(guildId, playlistName, id.String(), title, author)
}

func ffmpegmp3Command() *exec.Cmd {
	return exec.Command(
		"ffmpeg",
		"-y",
		"-i",
		TMPFILE_M4A,
		TMPFILE_MP3,
	)
}
func ytdlCommand(source string) *exec.Cmd {
	return exec.Command(
		"yt-dlp",
		"--sponsorblock-remove",
		"sponsor,music_offtopic,selfpromo,interaction,intro,outro,preview",
		"-f",
		"m4a",
		"-o",
		TMPFILE_M4A,
		source,
	)

}

func (e *ExecYouTubeDL) GetTitle(source string) (string, error) {
	cmd := exec.Command(
		"yt-dlp",
		"--print",
		"title",
		source,
	)

	b, err := cmd.Output()

	return strings.TrimSpace(string(b)), err

}

func (e *ExecYouTubeDL) GetAuthor(source string) (string, error) {
	cmd := exec.Command(
		"yt-dlp",
		"--print",
		"filename",
		"-o",
		"%(uploader)s",
		source,
	)

	b, err := cmd.Output()

	return strings.TrimSpace(string(b)), err
}

func (e *ExecYouTubeDL) GetThumbnail(source string) (string, error) {
	cmd := exec.Command(
		"yt-dlp",
		"--print",
		"thumbnail",
		"-o",
		"%(uploader)s",
		source,
	)

	b, err := cmd.Output()

	return strings.TrimSpace(string(b)), err
}

func fetchThumbnail(source string) error {
	resp, err := http.Get(source)
	if err != nil {
		return err
	}

	im, _, err := image.Decode(resp.Body)
	if err != nil {
		return err
	}

	cropped := imaging.Fill(im, 256, 256, imaging.Center, imaging.Lanczos)

	os.Remove(TMPFILE_ART)
	f, err := os.Create(TMPFILE_ART)
	if err != nil {
		return err
	}
	defer f.Close()
	defer resp.Body.Close()

	err = jpeg.Encode(f, cropped, nil)

	return err
}

func convertToDca() error {
	cmd := exec.Command(
		"dca",
		TMPFILE_M4A,
		TMPFILE_DCA,
	)

	err := cmd.Run()
	return err

}

func idToUrl(id string) string {
	return strings.Join([]string{
		"https://youtube.com/watch?v=",
		id,
	}, "")
}
