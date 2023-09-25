package youtubedl

import "go.uber.org/fx"

var Module = fx.Module("youtubedl",
	fx.Provide(
		NewYouTubeDL,
	),
)
