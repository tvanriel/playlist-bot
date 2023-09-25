package musicstore

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Configuration struct {
	Bucket string
}

type NewMusicStoreParams struct {
	fx.In

	S3     *minio.Client
	Config *Configuration
	Log    *zap.Logger
}

const METADATA_FILENAME = "metadata.json"
const MP3_FILENAME = "audio.mp3"
const M4A_FILENAME = "audio.m4a"
const DCA_FILENAME = "audio.dca"
const ART_FILENAME = "art.jpg"

func NewMusicStore(p NewMusicStoreParams) *MusicStore {
	return &MusicStore{
		S3:     p.S3,
		Bucket: p.Config.Bucket,
		Log:    p.Log,
	}
}

type MusicStore struct {
	S3     *minio.Client
	Bucket string
	Log    *zap.Logger
}

type Track struct {
	AlbumArt string
	Artist   string
	Name     string
	Source   string
	GuildID  string
	Uuid     string
	Formats  []Format
}

type Format struct {
	Name     string
	Filename string
}

func (t *Track) BaseName(with string) string {
	var sb strings.Builder
	sb.WriteString(t.GuildID)
	sb.WriteString("/")
	sb.WriteString(t.Uuid)
	sb.WriteString("/")
	sb.WriteString(with)
	return sb.String()
}

func (t *Track) ObjectName() string {
	return t.BaseName(METADATA_FILENAME)
}

func (t *Track) MP3() string {
	return t.BaseName(MP3_FILENAME)
}

func (t *Track) M4A() string {
	return t.BaseName(M4A_FILENAME)
}

func (t *Track) DCA() string {
	return t.BaseName(DCA_FILENAME)
}

func (t *Track) Art() string {
	return t.BaseName(ART_FILENAME)
}

func (m *MusicStore) Save(t *Track) error {
	obj, err := json.Marshal(t)
	if err != nil {
		return err
	}
	_, err = m.S3.PutObject(
		context.Background(),
		m.Bucket,
		t.ObjectName(),
		bytes.NewBuffer(obj),
		int64(len(obj)),
		minio.PutObjectOptions{},
	)
	return err

}
func (m *MusicStore) Find(guildId, uuid string) (*Track, error) {
	t := &Track{GuildID: guildId, Uuid: uuid}
	obj, err := m.S3.GetObject(context.Background(), m.Bucket, t.ObjectName(), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	read, err := io.ReadAll(obj)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(read, t)
	if err != nil {
		return nil, err
	}

	return t, err
}

func (m *MusicStore) UploadMP3(t *Track, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	_, err = m.S3.PutObject(context.Background(), m.Bucket, t.MP3(), f, -1, minio.PutObjectOptions{})
	if err == nil {
		t.Formats = append(t.Formats, Format{Name: "mp3", Filename: t.MP3()})
	}
	return err
}
func (m *MusicStore) UploadM4A(t *Track, path string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	_, err = m.S3.PutObject(context.Background(), m.Bucket, t.M4A(), f, -1, minio.PutObjectOptions{})
	if err == nil {
		t.Formats = append(t.Formats, Format{Name: "m4a", Filename: t.M4A()})
	}
	return err
}
func (m *MusicStore) UploadDCA(t *Track, path string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	_, err = m.S3.PutObject(context.Background(), m.Bucket, t.DCA(), f, -1, minio.PutObjectOptions{})
	if err == nil {
		t.Formats = append(t.Formats, Format{Name: "dca", Filename: t.DCA()})
	}
	return err
}

func (m *MusicStore) UploadAlbumArt(t *Track, path string) error {

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	_, err = m.S3.PutObject(context.Background(), m.Bucket, t.Art(), f, -1, minio.PutObjectOptions{})
	return err
}

func (m *MusicStore) GetDCA(t *Track) (io.ReadCloser, error) {
	obj, err := m.S3.GetObject(context.Background(), m.Bucket, t.DCA(), minio.GetObjectOptions{})
	return obj, err
}
func (m *MusicStore) GetAlbumArt(t *Track) (io.ReadCloser, error) {
	return m.S3.GetObject(context.Background(), m.Bucket, t.Art(), minio.GetObjectOptions{})
}
