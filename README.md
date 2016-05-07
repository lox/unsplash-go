
Unsplash API for Golang
=======================

Live coded at a Golang Melb Hack night.

Only `GET /users/:username/photos` is implemented.

## Example

Create a new [Unsplash Developer Application](https://unsplash.com/developers) and use the client id below.

```go
c := unsplash.NewClient("client id goes here")
photos, err := c.GetUserPhotos("lox")
if err != nil {
	log.Fatal(err)
}

for _, photo := range photos {
	log.Println(photo.Links.Download)
}
```

## Using the CLI

```bash
go get github.com/lox/unsplash-go/cli/unsplash
unsplash -user nasa -download ~/Documents/Wallpaper/Unsplash/NASA
```