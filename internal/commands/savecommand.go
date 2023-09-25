package commands

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	youtube "github.com/kkdai/youtube/v2"
	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"github.com/mitaka8/playlist-bot/internal/youtubedl"
	"go.uber.org/fx"
)

type NewSaveCommandParams struct {
	fx.In

	MusicStore    *musicstore.MusicStore
	YouTubeDL     youtubedl.YoutubeDL
	PlaylistStore *playliststore.PlaylistStore
}

func NewSaveCommand(p NewSaveCommandParams) *SaveCommand {
	return &SaveCommand{
		MusicStore:    p.MusicStore,
		YouTubeDL:     p.YouTubeDL,
		PlaylistStore: p.PlaylistStore,
	}
}

type SaveCommand struct {
	YouTubeDL     youtubedl.YoutubeDL
	MusicStore    *musicstore.MusicStore
	PlaylistStore *playliststore.PlaylistStore
}

var _ executor.Command = &SaveCommand{}

func (s *SaveCommand) SkipsPrefix() bool {
	return false
}

func (s *SaveCommand) Name() string {
	return "save"
}

var playlistRegex = regexp.MustCompile(`^https://(www\.?)(youtube\..+)/playlist\?list=[a-zA-Z0-9_-]+$`)
var videoRegex = regexp.MustCompile(`^https://(www\.?)(youtube.+)/watch\?v=[a-zA-Z0-9_-]+`)

func (s *SaveCommand) Apply(ctx *executor.Context) error {
	if len(ctx.Args) != 2 {
		_, err := ctx.Reply("usage: save <playlist-name> <url>")
		return err
	}
	playlistName := ctx.Args[0]

	guildId := ctx.Message.GuildID

	exists, err := s.PlaylistStore.PlaylistExists(guildId, playlistName)

	if err != nil {
		return err
	}
	if !exists {
		ctx.Reply("That playlist does not exist. You must first create it!")
		return nil
	}

	url := ctx.Args[1]

	if playlistRegex.MatchString(url) {
		ytclient := youtube.Client{}
		playlist, err := ytclient.GetPlaylist(url)
		if err != nil {
			return err
		}

		for i := range playlist.Videos {
			id, err := uuid.NewRandom()

			if err != nil {
				ctx.Error(err)
				continue

			}
			s.YouTubeDL.Save(
				idToUrl(playlist.Videos[i].ID),
				guildId,
				id.String(),
			)
			s.PlaylistStore.Append(guildId, playlistName, id.String())
			ctx.Reply(savingText(playlist.Videos[i].ID))
		}

	} else if videoRegex.MatchString(url) {
		id, err := uuid.NewRandom()

		if err != nil {
			return err
		}
		s.YouTubeDL.Save(
			url,
			guildId,
			id.String(),
		)
		s.PlaylistStore.Append(guildId, playlistName, id.String())
		ctx.Reply("Saving...")

	} else {

		ctx.Reply("Cannot detect URL format, give me a youtube video url or a playlist url")
		return nil
	}

	return nil
}

func idToUrl(id string) string {
	return strings.Join([]string{
		"https://youtube.com/watch?v=",
		id,
	}, "")
}
func savingText(id string) string {
	return strings.Join([]string{
		"Saving ",
		id,
		"...",
	}, "")
}
