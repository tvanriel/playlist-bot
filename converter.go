package playlistdiscordbot

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jonas747/dca"
	"github.com/kkdai/youtube/v2"
)

func DownloadAndConvert(url, outFile string, onError func(error), onSuccess func()) {
	downloadTemp(url, onError, func(tmpFile string) {
		convertFile(tmpFile, outFile, onError, onSuccess)
	})
}
func YoutubeDownloadAndConvert(url, outFile string, onError func(error), onSuccess func()) {
	downloadYoutubeTemp(url, onError, func(tmpFile string) {
		convertFile(tmpFile, outFile, onError, onSuccess)
	})
}

func downloadTemp(url string, onError func(error), onSuccess func(string)) {
	resp, err := http.Get(url)
	if err != nil {
		onError(err)
		return
	}
	file, err := os.CreateTemp(os.TempDir(), "playlist-download")
	if err != nil {
		onError(err)
		return
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		onError(err)
		return
	}

	file.Close()
	onSuccess(file.Name())
	os.Remove(file.Name())
}

func convertFile(tmpFile string, outfile string, onError func(error), onSuccess func()) {

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	encodingSession, err := dca.EncodeFile(tmpFile, options)
	if err != nil {
		onError(err)
		return
	}
	defer encodingSession.Cleanup()

	output, err := os.Create(outfile)
	if err != nil {
		onError(err)
		return
	}

	_, err = io.Copy(output, encodingSession)
	if err != nil {
		log.Printf("Err %v\n", err)
		onError(err)
		return
	}
	output.Close()
	onSuccess()
}

func downloadYoutubeTemp(url string, onError func(error), onSuccess func(string)) {
	client := youtube.Client{}
	videoInfo, err := client.GetVideo(url)
	if err != nil {
		onError(err)
		return
	}

	format := videoInfo.Formats.WithAudioChannels()
	stream, _, err := client.GetStream(videoInfo, &format[0])
	if err != nil {
		onError(err)
		return
	}

	file, err := os.CreateTemp(os.TempDir(), "playlist-download*.mp4")
	defer os.Remove(file.Name())

	if err != nil {
		onError(err)
		return
	}

	_, err = io.Copy(file, stream)
	if err != nil {
		onError(err)
		return
	}
	file.Close()
	onSuccess(file.Name())

}
