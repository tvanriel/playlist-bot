package web

import (
	"bytes"
	"errors"
	"io"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mitaka8/playlist-bot/internal/guildstore"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"github.com/mitaka8/playlist-bot/internal/progresstracker"
	"github.com/tvanriel/cloudsdk/http"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

type NewWebInterfaceParams struct {
	fx.In

	Log             *zap.Logger
	GuildRepository *guildstore.GuildRepository
	Templates       *Templater
	ProgressTracker *progresstracker.ProgressTracker
	MusicStore      *musicstore.MusicStore
	PlaylistStore   *playliststore.PlaylistStore
}

func NewWebInterface(p NewWebInterfaceParams) *WebInterface {
	return &WebInterface{
		Log:             p.Log.Named("web"),
		GuildRepository: p.GuildRepository,
		templates:       p.Templates,
		ProgressTracker: p.ProgressTracker,
		MusicStore:      p.MusicStore,
		PlaylistStore:   p.PlaylistStore,
	}
}

var _ http.RouteGroup = &WebInterface{}

type WebInterface struct {
	GuildRepository *guildstore.GuildRepository
	Log             *zap.Logger
	templates       *Templater
	ProgressTracker *progresstracker.ProgressTracker
	PlaylistStore   *playliststore.PlaylistStore

	MusicStore *musicstore.MusicStore
}

func (w *WebInterface) ApiGroup() string { return "" }
func (w *WebInterface) Version() string  { return "" }

func (w *WebInterface) Handler(e *echo.Group) {

	e.GET("", func(c echo.Context) error {

		guilds, err := w.GuildRepository.GetGuilds()
		if err != nil {
			return RenderError(c, err)
		}

		return c.Render(200, "index.tpl.html", indexHtmlParams{Guilds: guilds})
	})
	e.Static("app", "web/static")

	e.GET("player/:guildId", func(c echo.Context) error {
		return c.Render(200, "guild.tpl.html", guildHtmlParams{

			Snowflake: c.Param("guildId"),
		})
	})

	e.GET("progress/:guildId", func(c echo.Context) error {
		guildId := c.Param("guildId")
		if guildId == "" {
			return RenderError(c, errors.New("guildId is required"))
		}
		websocket.Handler(func(ws *websocket.Conn) {
			defer ws.Close()

			ch, err := w.ProgressTracker.Consume(guildId)
			if err != nil {
				w.Log.Error("cannot listen for progress", zap.Error(err), zap.String("guildId", guildId))
				ws.Close()
				return
			}
			for m := range ch {
				buf := bytes.NewBufferString("")
				progress := int64(float32(m.Current) / float32(m.Max) * 100.0)
				t, err := w.MusicStore.Find(guildId, m.Track)
				if err != nil {
					return
				}

				playlistId, err := w.GuildRepository.GetPlaying(guildId)
				if err != nil {
					return
				}
				playlist, err := w.PlaylistStore.FindByID(playlistId)
				if err != nil {
					return
				}

				err = w.templates.Templates.Lookup("progress.tpl.html").Execute(buf, progressHtmlParams{
					Track:              t.Name,
					GuildId:            guildId,
					TrackId:            t.Uuid,
					Artist:             t.Artist,
					Playlist:           playlist.Name,
					ProgressInteger:    strconv.FormatInt(progress, 10),
					ProgressPercentage: strconv.FormatInt(progress, 10) + "%",
				})
				if err != nil {
					w.Log.Error("error rendering template to buffer", zap.Error(err))
					continue
				}
				_, err = io.Copy(ws, buf)
				if err != nil {
					w.Log.Error("error copying template to websocket", zap.Error(err))
					return
				}
			}

		}).ServeHTTP(c.Response(), c.Request())
		return nil
	})
	e.GET("art/:guildId/:trackId", func(c echo.Context) error {
		t, err := w.MusicStore.Find(c.Param("guildId"), c.Param("trackId"))
		if err != nil {
			return err
		}

		reader, err := w.MusicStore.GetAlbumArt(t)
		if err != nil {
			return err
		}
		_, err = io.Copy(c.Response().Writer, reader)
		return err

	})
}

type indexHtmlParams struct {
	Guilds []guildstore.GuildModel
}

type guildHtmlParams struct {
	Snowflake string
}

type progressHtmlParams struct {
	Track              string
	GuildId            string
	TrackId            string
	Artist             string
	Playlist           string
	ProgressInteger    string
	ProgressPercentage string
}
