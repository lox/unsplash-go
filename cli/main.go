package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/lox/unsplash-go"
)

func main() {
	var (
		clientId = flag.String("clientid", "", "The client id for the unsplash application")
		user     = flag.String("user", "", "The user to download photos from")
	)

	flag.Parse()

	c := unsplash.NewClient(*clientId)
	photos, err := c.GetUserPhotos(*user)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	for _, photo := range photos {
		log.Printf("Downloading %s (file=%s.jpg)", photo.Links.Download, photo.ID)

		f, err := os.Create(photo.ID + ".jpg")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		wg.Add(1)
		go func(photo unsplash.Photo) {
			resp, err := http.Get(photo.Links.Download)
			if err != nil {
				log.Fatal(err)
			}

			io.Copy(f, resp.Body)
			defer resp.Body.Close()

			log.Printf("Finished downloading %s.jpg", photo.ID)
			wg.Done()
		}(photo)
	}

	wg.Wait()
	log.Printf("All done!")
}
