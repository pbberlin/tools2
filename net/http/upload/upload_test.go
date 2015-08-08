package upload

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
)

func TestUpload(t *testing.T) {

	log.SetFlags(log.Lshortfile)

	files := []string{"test.zip", "test.jpg"}

	for _, v := range files {

		curdir, _ := os.Getwd()
		filePath := path.Join(curdir, v)

		extraParams := map[string]string{
			"getparam1":   "val1",
			"description": "A zip file - containing dirs and files",
		}

		urlUp := "https://google.com/upload"
		urlUp = "http://localhost:8085" + UrlUploadReceive
		// urlUp = "https://libertarian-islands.appspot.com" + UrlUploadReceive

		request, err := CreateFilePostRequest(
			urlUp, "filefield", filePath, extraParams)
		if err != nil {
			log.Fatal(err)
		}

		client := &http.Client{}
		log.Printf("Sending req ... %v", urlUp)
		resp, err := client.Do(request)

		if err != nil {
			log.Fatal(err)
		} else {
			bts, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()
			for k, v := range resp.Header {
				fmt.Println("\tHdr: ", k, v)
			}
			fmt.Printf("status: %v\n", resp.StatusCode)

			bod := string(bts)
			bods := strings.Split(bod, "<span class='body'></span>")
			if len(bods) == 3 {
				bod = bods[1]
			}

			fmt.Printf("body:   %s\n", bod)
		}

	}
}
