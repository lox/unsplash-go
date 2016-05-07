
Unsplash API for Golang
=======================

Live coded at a Golang Melb Hack night.

## Implemented

 - Authentication via OAuth
 - `GET /users/:username/photos`

## Example

Create a new [Unsplash Developer Application](https://unsplash.com/developers) and use the client id below, or for how to use OAuth, see [cli/unsplash/oauth.go](cli/unsplash/oauth.go)

```go
c := unsplash.NewPublicClient("client id from above goes here")
photos, err := c.GetUserPhotos("lox", func(p unsplash.Photo) error {
	log.Println(photo.Links.Download)
	return nil
})
if err != nil {
	log.Fatal(err)
}
```

## Using the CLI

```bash
go get github.com/lox/unsplash-go/cli/unsplash
unsplash -user nasa -download ~/Documents/Wallpaper/Unsplash/NASA
```