package playliststore

import (
	"go.uber.org/fx"
)

var Module = fx.Module("playliststore",

	fx.Provide(
		NewMySQLPlaylistStore,
	),
	fx.Invoke(MigratePlaylistStore),
)
