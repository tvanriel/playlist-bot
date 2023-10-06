package commands

import (
	executor "github.com/mitaka8/playlist-bot/internal/command-executor"
	"go.uber.org/fx"
)

var Module = fx.Module("commands", fx.Provide(
	executor.AsCommand(NewAddPlaylistCommand),
	executor.AsCommand(NewJoinCommand),
	executor.AsCommand(NewListPlaylistCommand),
	executor.AsCommand(NewPlayCommand),
	executor.AsCommand(NewSaveCommand),
	executor.AsCommand(NewSearchCommand),
))
