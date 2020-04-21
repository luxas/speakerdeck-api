# speakerdeck-api

A scraper implementation for SpeakerDeck, as a workaround due to the lack of an official API.

You can either use this as a library (`import speakerdeck "github.com/luxas/speakerdeck-api"`) or the
ready-made (very lightweight) API implementation under `cmd/speakerdeck-api`.

## API Usage

Install and start the API:

```console
$ go get github.com/luxas/speakerdeck-api/cmd/speakerdeck-api
$ $GOPATH/bin/speakerdeck-api
INFO[0000] Starting Speakerdeck API...
```

Get information about a user:

```shell
curl http://localhost:8080/api/users/luxas
```

```json
{
  "author": {
    "name": "Lucas Käldström",
    "handle": "luxas",
    "link": "https://speakerdeck.com/luxas",
    "avatarLink": "https://secure.gravatar.com/avatar/111ac0b31c0dc219c84ddadedc8e5f67?s=128"
  },
  "abstract": "Lucas is a cloud native enthusiast who has been serving the Kubernetes \u0026 CNCF communities in lead positions for more than 4 years. Lucas is a CNCF Ambassador, running 3 meetup groups in Finland and coordinating the Cloud Native Nordics meetups. He got the \"Top Cloud Native Ambassador\" award together with Sarah Novotny 2017. Lucas has e.g. shepherded kubeadm from inception to GA as a co-lead for SIG Cluster Lifecycle, ported Kubernetes to Raspberry Pi and multiple other platforms. Lucas is a CKA, runs a consulting company \"luxas labs\" for cloud native tech and has spoken at six KubeCons.",
  "talkPreviews": [
    ...,
    {
      "title": "Getting Started in the Kubernetes Community",
      "id": "getting-started-in-the-kubernetes-community",
      "views": 92,
      "stars": 1,
      "date": "2019-05-21T00:00:00Z",
      "link": "https://speakerdeck.com/luxas/getting-started-in-the-kubernetes-community",
      "dataID": "6816e10f104a44cebb0915b392cadd2d"
    },
    ...,
  ]
}
```

Get detailed information about a user's talks (all of them):

```shell
curl http://localhost:8080/api/talks/luxas
```

```json
[
  ...,
  {
    "title": "Getting Started in the Kubernetes Community",
    "id": "getting-started-in-the-kubernetes-community",
    "views": 92,
    "stars": 1,
    "date": "2019-05-21T00:00:00Z",
    "link": "https://speakerdeck.com/luxas/getting-started-in-the-kubernetes-community",
    "dataID": "6816e10f104a44cebb0915b392cadd2d",
    "author": {
      "name": "Lucas Käldström",
      "handle": "luxas",
      "link": "https://speakerdeck.com/luxas",
      "avatarLink": "https://secure.gravatar.com/avatar/111ac0b31c0dc219c84ddadedc8e5f67?s=47"
    },
    "category": "Technology",
    "categoryLink": "https://speakerdeck.com/c/technology",
    "downloadLink": "https://speakerd.s3.amazonaws.com/presentations/6816e10f104a44cebb0915b392cadd2d/Lucas_Kaldstrom-Nikhita_Raghunath_-_May_21_-_Morning.pdf",
    "extraLinks": {
      "github.com": [
        "https://github.com/nikhita"
      ]
    },
    "hide": false
  },
  ...,
]
```

Get detailed information about one of the user's talks:

```shell
curl http://localhost:8080/api/talks/luxas/getting-started-in-the-kubernetes-community
```

```json
[
  {
    "title": "Getting Started in the Kubernetes Community",
    "id": "getting-started-in-the-kubernetes-community",
    "views": 92,
    "stars": 1,
    "date": "2019-05-21T00:00:00Z",
    "link": "https://speakerdeck.com/luxas/getting-started-in-the-kubernetes-community",
    "dataID": "6816e10f104a44cebb0915b392cadd2d",
    "author": {
      "name": "Lucas Käldström",
      "handle": "luxas",
      "link": "https://speakerdeck.com/luxas",
      "avatarLink": "https://secure.gravatar.com/avatar/111ac0b31c0dc219c84ddadedc8e5f67?s=47"
    },
    "category": "Technology",
    "categoryLink": "https://speakerdeck.com/c/technology",
    "downloadLink": "https://speakerd.s3.amazonaws.com/presentations/6816e10f104a44cebb0915b392cadd2d/Lucas_Kaldstrom-Nikhita_Raghunath_-_May_21_-_Morning.pdf",
    "extraLinks": {
      "github.com": [
        "https://github.com/nikhita"
      ]
    },
    "hide": false
  }
]
```

For reference you can visit [https://speakerdeck.com/luxas/getting-started-in-the-kubernetes-community](https://speakerdeck.com/luxas/getting-started-in-the-kubernetes-community) to check where the data is coming from.

### Geolocation

`speakerdeck-api` also has support for extensions, the extension that currently exists is `LocationExtension` (in `./location`), which
can using a Google Maps API key (with access to the Geocoding API) geolocate your talks just by you putting `Location: <address>` in the Speakerdeck talk description!

Enable it by giving the Google Maps key to the API:

```shell
$GOPATH/bin/speakerdeck-api -maps-api-key <API_KEY>
```

Now, notice the `location` field which is generated based on the `Location: Fira Gran Via, Av. Joan Carles I, Barcelona, Spain` string embedded in the talk description:

```shell
curl http://localhost:8080/api/talks/luxas/getting-started-in-the-kubernetes-community
```

```json
[
  {
    "title": "Getting Started in the Kubernetes Community",
    "id": "getting-started-in-the-kubernetes-community",
    "views": 92,
    "stars": 1,
    "date": "2019-05-21T00:00:00Z",
    "link": "https://speakerdeck.com/luxas/getting-started-in-the-kubernetes-community",
    "dataID": "6816e10f104a44cebb0915b392cadd2d",
    "author": {
      "name": "Lucas Käldström",
      "handle": "luxas",
      "link": "https://speakerdeck.com/luxas",
      "avatarLink": "https://secure.gravatar.com/avatar/111ac0b31c0dc219c84ddadedc8e5f67?s=47"
    },
    "category": "Technology",
    "categoryLink": "https://speakerdeck.com/c/technology",
    "downloadLink": "https://speakerd.s3.amazonaws.com/presentations/6816e10f104a44cebb0915b392cadd2d/Lucas_Kaldstrom-Nikhita_Raghunath_-_May_21_-_Morning.pdf",
    "extraLinks": {
      "github.com": [
        "https://github.com/nikhita"
      ]
    },
    "hide": false,
    "location": {
      "requestedAddress": "Fira Gran Via, Av. Joan Carles I, Barcelona, Spain",
      "resolvedAddress": "Av. Joan Carles I, 64, 08908 L'Hospitalet de Llobregat, Barcelona, Spain",
      "lat": 41.3546878,
      "lng": 2.1277339
    }
  }
]
```

## Library Usage

Check out the documentation on [Godoc](https://godoc.org/github.com/luxas/speakerdeck-api) or [pkg.go.dev](https://pkg.go.dev/github.com/luxas/speakerdeck-api)!

## License

[MIT](LICENSE)
