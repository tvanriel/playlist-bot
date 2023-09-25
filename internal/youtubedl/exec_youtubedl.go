package youtubedl

import (
	_ "golang.org/x/image/webp"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"go.uber.org/zap"
)

type ExecYouTubeDL struct {
	Configuration Configuration
	S3            *minio.Client
	MusicStore    *musicstore.MusicStore
	Log           *zap.Logger
}

func (e *ExecYouTubeDL) Save(source string, guildId string, uuid string) {
	e.Log.Info("downloading",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	ytdl := ytdlCommand(source)
	ffmpegmp3 := ffmpegmp3Command()

	e.Log.Info("downloading youtube file",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	output, err := ytdl.Output()
	if err != nil {
		e.Log.Error("cannot fetch youtube file", zap.Error(err), zap.String("output", string(output)))
		return
	}
	e.Log.Info("convert to mp3",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	output, err = ffmpegmp3.Output()
	if err != nil {
		e.Log.Error("cannot convert to mp3", zap.Error(err), zap.String("output", string(output)))
		return
	}
	e.Log.Info("convert to dca",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	err = convertToDca()
	if err != nil {
		e.Log.Error("cannot convert to dca", zap.Error(err))
		return
	}
	e.Log.Info("fetch title",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	title, err := e.GetTitle(source)
	if err != nil {
		e.Log.Error("cannot fetch title", zap.Error(err))
		return
	}
	e.Log.Info("fetch author",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	author, err := e.GetAuthor(source)
	if err != nil {
		e.Log.Error("cannot fetch author", zap.Error(err))
		return
	}
	e.Log.Info("fetch thumbnail url",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	thumbnail, err := e.GetThumbnail(source)
	if err != nil {
		e.Log.Error("cannot fetch AlbumArt", zap.Error(err))
		return
	}
	t := &musicstore.Track{
		AlbumArt: thumbnail,
		Name:     title,
		Artist:   author,
		Source:   source,
		GuildID:  guildId,
		Uuid:     uuid,
	}
	e.Log.Info("downloading thumbnail file",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	err = fetchThumbnail(thumbnail)
	if err != nil {
		e.Log.Error("Failed to download AlbumArt", zap.Error(err))
		return
	}
	e.Log.Info("uploading mp3 to bucket",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	err = e.MusicStore.UploadMP3(t, "/tmp/audio.mp3")
	if err != nil {
		e.Log.Error("Failed to upload mp3", zap.Error(err))
		return
	}
	e.Log.Info("uploading m4a to bucket",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	err = e.MusicStore.UploadM4A(t, "/tmp/audio.m4a")
	if err != nil {
		e.Log.Error("Failed to upload m4a", zap.Error(err))
		return
	}
	e.Log.Info("uploading dca to bucket",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	err = e.MusicStore.UploadDCA(t, "/tmp/audio.dca")
	if err != nil {
		e.Log.Error("Failed to upload dca", zap.Error(err))
		return
	}

	e.Log.Info("uploading albumart to bucket",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	err = e.MusicStore.UploadAlbumArt(t, "/tmp/art.jpg")
	if err != nil {
		e.Log.Error("Failed to upload albumart", zap.Error(err))
		return
	}

	e.Log.Info("uploading manifest to bucket",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)

	err = e.MusicStore.Save(t)
	if err != nil {
		e.Log.Error("Failed to save track", zap.Error(err))
		return
	}

	e.Log.Info("done",
		zap.String("source", source),
		zap.String("guildId", guildId),
		zap.String("uuid", uuid),
	)


        os.Remove("/tmp/audio.m4a")
        os.Remove("/tmp/audio.mp3")
        os.Remove("/tmp/audio.dca")
        os.Remove("/tmp/art.jpg")

}

func ffmpegmp3Command() *exec.Cmd {
	return exec.Command(
		"ffmpeg",
		"-y",
		"-i",
		"/tmp/audio.m4a",
		"/tmp/audio.mp3",
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
		"/tmp/audio.m4a",
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

	os.Remove("/tmp/art.jpg")
	f, err := os.Create("/tmp/art.jpg")
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
		"/tmp/audio.m4a",
		"/tmp/audio.dca",
	)

	err := cmd.Run()
	return err

}
