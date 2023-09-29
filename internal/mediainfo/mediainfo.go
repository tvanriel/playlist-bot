package mediainfo

import (
	"os/exec"

	"go.uber.org/fx"
)

type NewMediaInfoQueryParams struct {
	fx.In

	Configuration Configuration
}

func NewMediaInfoQuery(p NewMediaInfoQueryParams) *MediaInfoQuery {
	return &MediaInfoQuery{
		Binary: p.Configuration.Location,
	}
}

type MediaInfoQuery struct {
	Binary string
}

type MediaInfoOutput struct {
	CreatingLibrary *MediaInfoOutputCreatingLibrary `json:"creatingLibrary"`
	Media           *MediaInfoOutputMedia           `json:"media"`
}

type MediaInfoOutputCreatingLibrary struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Url     string `json:"url"`
}

type MediaInfoOutputMedia struct {
	AtRef string                      `json:"@ref"`
	Track []MediaInfoOutputMediaTrack `json:"track"`
}

type MediaInfoOutputMediaTrack struct {
	AtType              string `json:"@type"`
	AudioCount          string `json:"AudioCount"`
	FileExtension       string `json:"FileExtension"`
	Format              string `json:"Format"`
	FileSize            string `json:"FileSize"`
	Duration            string `json:"Duration"`
	OverallBitRate_Mode string `json:"OverallBitRate_Mode"`
}

func (m *MediaInfoQuery) Query(path string) (*MediaInfoOutput, error) {
	cmd := exec.Command(m.Binary, "--output=JSON", path)

}
