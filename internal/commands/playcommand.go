package commands

import (
	"strings"

	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"github.com/mitaka8/playlist-bot/internal/guildstore"
	"github.com/mitaka8/playlist-bot/internal/playliststore"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type NewPlayCommandParams struct {
	fx.In

	Log                *zap.Logger
	GuildRepository    *guildstore.GuildRepository
	PlaylistRepository *playliststore.PlaylistStore
}

func NewPlayCommand(p NewPlayCommandParams) *PlayCommand {
	return &PlayCommand{
		Log:                p.Log.Named("play"),
		GuildRepository:    p.GuildRepository,
		PlaylistRepository: p.PlaylistRepository,
	}
}

type PlayCommand struct {
	Log                *zap.Logger
	GuildRepository    *guildstore.GuildRepository
	PlaylistRepository *playliststore.PlaylistStore
}

var _ executor.Command = &PlayCommand{}

func (p *PlayCommand) Name() string {
	return "play"
}
func (p *PlayCommand) SkipsPrefix() bool { return false }
func (p *PlayCommand) Apply(ctx *executor.Context) error {
        log := p.Log.With(ctx.ZapFields()...)

        log.Info("Play")

	if len(ctx.Args) != 1 {
		ctx.Reply("See `list` for a list of playlists\nusage: play <name>")
	}
	playlist, err := p.PlaylistRepository.FindByGuildAndName(ctx.Message.GuildID, ctx.Args[0])
	if err != nil {
                log.Warn("Failed to find playlist", zap.Error(err))
		return err
	}
	err = p.GuildRepository.SetPlaying(ctx.Message.GuildID, playlist.ID)
	if err != nil {
                log.Warn("Failed to set guild.currently_playing", zap.Error(err))
		return err
	}

        log.Warn("guild is now playing", zap.String("playlist", playlist.Name))
	ctx.Reply(nowPlaying(playlist.Name))
	return nil
}
func nowPlaying(name string) string {
	return strings.Join([]string{
		"Now playing: ",
		name,
	}, "")
}
