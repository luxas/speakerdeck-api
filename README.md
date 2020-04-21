# speakerdeck-scraper

A scraper implementation for SpeakerDeck, as a workaround due to the lack of an official API.

You can either use this as a library (`import speakerdeck "github.com/luxas/speakerdeck-scraper"`) or the
ready-made (very lightweight) API implementation under `cmd/speakerdeck-api`.

`speakerdeck-scraper` also has support for extensions, the extension that currently exists is Location, which
can using a Google Maps API key (with access to the Geocoding API) geolocate your talks just by you putting `Location: <address>` in the Speakerdeck talk description!

## API Usage

```console
$ go get github.com/luxas/speakerdeck-scraper/cmd/speakerdeck-api
$ $GOPATH/bin/speakerdeck-api
INFO[0000] Starting Speakerdeck API...
$ curl http://localhost:8080/api/users/luxas
{ ... }
$ curl http://localhost:8080/api/talks/luxas
[ { ... }, { ... }, ... ]
$ curl http://localhost:8080/api/talks/luxas/getting-started-in-the-kubernetes-community
{ ... }
```

## Library Usage

Check out the Godoc!
More info here coming soon...

## License

[MIT](LICENSE)
