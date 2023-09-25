package guildstore

import (
	"go.uber.org/fx"
)

var Module = fx.Module("guildstore",
	fx.Provide(
		NewGuildRepository,
	),
	fx.Invoke(MigrateGuildRepository),
)
