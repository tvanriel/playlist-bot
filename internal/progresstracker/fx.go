package progresstracker

import "go.uber.org/fx"

var Module = fx.Module("progresstracker",

	fx.Provide(
		NewProgressTracker,
	),
)

func StartReporting(p *ProgressTracker) {
	p.Publish()
}
