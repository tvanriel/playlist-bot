package commands

import (
	"strings"

	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/musicstore"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"github.com/mitaka8/playlist-bot/internal/youtubedl"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NewSaveCommandParams struct {
	fx.In

	MusicStore    *musicstore.MusicStore
	YouTubeDL     youtubedl.YoutubeDL
	PlaylistStore *playliststore.PlaylistStore
        Log *zap.Logger
}

func NewSaveCommand(p NewSaveCommandParams) *SaveCommand {
	return &SaveCommand{
		MusicStore:    p.MusicStore,
		YouTubeDL:     p.YouTubeDL,
		PlaylistStore: p.PlaylistStore,
                Log:           p.Log.Named("save"),
	}
}

type SaveCommand struct {
	YouTubeDL     youtubedl.YoutubeDL
	MusicStore    *musicstore.MusicStore
	PlaylistStore *playliststore.PlaylistStore
        Log *zap.Logger
}

var _ executor.Command = &SaveCommand{}

func (s *SaveCommand) SkipsPrefix() bool {
	return false
}

func (s *SaveCommand) Name() string {
	return "save"
}

func (s *SaveCommand) Apply(ctx *executor.Context) error {
        log := s.Log.With(ctx.ZapFields()...)
	if len(ctx.Args) != 2 {
		_, err := ctx.Reply("usage: save <playlist-name> <url>")
		return err
	}
	playlistName := ctx.Args[0]
	guildId := ctx.Message.GuildID
	exists, err := s.PlaylistStore.PlaylistExists(guildId, playlistName)

	if err != nil {
                log.Info("Saving...")
		return err
	}

	if !exists {
		ctx.Reply("That playlist does not exist. You must first create it!")
		return nil
	}

	url := ctx.Args[1]

	ctx.Reply("Saving...")

        params := youtubedl.YouTubeDLParams{
		Source:       url,
		GuildID:      guildId,
		PlaylistName: playlistName,
	}
        log.With(params.ZapFields()...).Info("Saving...")
	err = s.YouTubeDL.Save(params)
	if err != nil {
                log.With(params.ZapFields()...).Error("Failed to save", )
		ctx.Error(err)
	}

	return nil
}

func savingText(id string) string {
	return strings.Join([]string{
		"Saving ",
		id,
		"...",
	}, "")
}
