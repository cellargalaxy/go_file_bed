package main

import (
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"github.com/cellargalaxy/go_common/util"
	"github.com/cellargalaxy/go_file_bed/config"
	"github.com/cellargalaxy/go_file_bed/corn"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func init() {
	logrus.SetLevel(config.Config.LogLevel)
	util.InitDefaultLog()
	corn.Init()
}

/**
export server_name=go_file_bed
export server_center_address=http://127.0.0.1:7557
export server_center_secret=secret_secret

server_name=go_file_bed;server_center_address=http://127.0.0.1:7557;server_center_secret=secret_secret
*/
func main() {
	//err := controller.Controller()
	//if err != nil {
	//	panic(err)
	//}

	// create client
	client := grab.NewClient()
	req, _ := grab.NewRequest("./aaa.apk", "http://127.0.0.1:8880/file/tmp/%E5%BC%B9%E5%BC%B9%E5%A5%87%E5%A6%99%E5%86%92%E9%99%A9_20210604104002.apk")

	// start download
	fmt.Printf("Downloading %v...\n", req.URL())
	resp := client.Do(req)
	fmt.Printf("  %v\n", resp.HTTPResponse.Status)

	// start UI loop
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()

Loop:
	for {
		select {
		case <-t.C:
			fmt.Printf("  transferred %v / %v bytes (%.2f%%)\n",
				resp.BytesComplete(),
				resp.Size(),
				100*resp.Progress())

		case <-resp.Done:
			// download is complete
			break Loop
		}
	}

	// check for errors
	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Download saved to ./%v \n", resp.Filename)

	// Output:
	// Downloading http://www.golang-book.com/public/pdf/gobook.pdf...
	//   200 OK
	//   transferred 42970 / 2893557 bytes (1.49%)
	//   transferred 1207474 / 2893557 bytes (41.73%)
	//   transferred 2758210 / 2893557 bytes (95.32%)
	// Download saved to ./gobook.pdf
}
