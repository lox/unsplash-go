package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/lox/unsplash-go"
)

var (
	user     = flag.String("user", "", "The user to download photos from")
	download = flag.String("download", "", "The directory to download photos to")
)

func main() {
	flag.Parse()

	oAuthClient, err := newOAuthClient()
	if err != nil {
		log.Fatal(err)
	}

	client := unsplash.Client{
		Client: oAuthClient,
	}

	var wg sync.WaitGroup

	err = client.GetUserPhotos(*user, func(photo unsplash.Photo) error {
		wg.Add(1)
		go func() {
			log.Printf("Downloading %s (file=%s.jpg)", photo.Links.Download, photo.ID)
			if err = downloadPhoto(*download, photo); err != nil {
				log.Fatal(err)
			}
			log.Printf("Finished downloading %s.jpg", photo.ID)
			wg.Done()
		}()

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()
	log.Printf("All done!")
}

func downloadPhoto(dir string, photo unsplash.Photo) error {
	absDir, err := filepath.Abs(*download)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(filepath.Join(absDir, photo.ID+".jpg"))
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(photo.Links.Download)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return err
	}

	return resp.Body.Close()
}

func cacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache")
	}
	return "."
}
