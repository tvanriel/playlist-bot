package musicstore

import "go.uber.org/fx"

var Module = fx.Module("musicstore",
	fx.Provide(
		NewMusicStore,
	),
)
