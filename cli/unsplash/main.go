package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/lox/unsplash-go"
)

const (
	parallel = 4
)

func main() {
	var (
		// sources
		user       = flag.String("user", "", "Search for a specific users photos")
		userlikes  = flag.String("userlikes", "", "Search for a specific users liked photos")
		collection = flag.String("collection", "", "Search a specific collection by id")

		// actions
		wallpaper = flag.Bool("wallpaper", false, "Set downloaded images as desktop wallpaper")

		// filters
		dir      = flag.String("dir", "", "The directory to download photos to")
		parallel = flag.Int("parallel", 4, "The number of photos to download in parallel")
		order    = flag.String("order", "latest", "How to order the photos")
		limit    = flag.Int64("limit", -1, "The maximum number of photos to load")
	)

	flag.Parse()

	if *user == "" && *userlikes == "" && *collection == "" {
		fmt.Println("Either user, userlikes or collection must be specified as a source")
		flag.Usage()
		os.Exit(1)
	}

	oAuthClient, err := newOAuthClient()
	if err != nil {
		log.Fatal(err)
	}

	client := unsplash.Client{
		Client: oAuthClient,
	}

	var photos = make(chan unsplash.Photo)
	var counter int64
	var photosFunc = func(photo unsplash.Photo) (bool, error) {
		if *limit > 0 && atomic.LoadInt64(&counter) >= *limit {
			return true, nil
		}
		atomic.AddInt64(&counter, 1)
		photos <- photo
		return false, nil
	}

	// search for photos
	go func() {
		defer close(photos)
		switch {
		case *user != "":
			client.GetUserPhotos(*user, *order, photosFunc)
		case *userlikes != "":
			client.GetUsersLikes(*userlikes, *order, photosFunc)
		case *collection != "":
			client.GetCollection(*collection, *order, photosFunc)
		}
	}()

	if *wallpaper {
		if *dir == "" {
			if err := os.MkdirAll(cacheDir(), 0700); err != nil {
				log.Fatal(err)
			}
			tmpDir, err := ioutil.TempDir(cacheDir(), "")
			if err != nil {
				log.Fatal(err)
			}
			*dir = tmpDir
		}
		log.Printf("Setting wallpaper dir to %s after download", *dir)
	}

	// download photos
	var wg sync.WaitGroup

	log.Printf("Downloading to %s, parallel=%d", *dir, *parallel)
	for i := 0; i < *parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for photo := range photos {
				log.Printf("Downloading %s (file=%s.jpg)", photo.Links.Download, photo.ID)
				if err = downloadPhoto(*dir, photo); err != nil {
					log.Fatal(err)
				}
				log.Printf("Finished downloading %s.jpg", photo.ID)
			}
		}()
	}

	wg.Wait()

	if *wallpaper {
		log.Printf("Setting wallpaper dir to %s", *dir)
		if err := setWallpaperDir(*dir); err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("All done!")
}

func setWallpaperDir(dir string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("Only darwin is supported for wallpaper setting")
	}

	command := fmt.Sprintf(`
	tell application "System Events"
	    tell current desktop
	        set picture rotation to 1 -- (0=off, 1=interval, 2=login, 3=sleep)
	        set random order to true
	        set pictures folder to POSIX file "%s"
	        set change interval to 86400 -- seconds
	    end tell
	end tell
	`, dir)

	return exec.Command("/usr/bin/osascript", "-e", command).Run()
}

func downloadPhoto(dir string, photo unsplash.Photo) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absDir, 0700); err != nil {
		return err
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

func downloadDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Pictures", "Unsplash")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), "unsplash")
	}
	return "."
}

func cacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Caches", "unsplash-go")
	case "linux", "freebsd":
		return filepath.Join(os.Getenv("HOME"), ".cache", "unsplash-go")
	}
	return os.TempDir()
}
