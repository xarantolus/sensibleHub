# sensibleHub
sensibleHub is a self-hosted music management server. It allows managing your music collection from any device (that has a web browser) 
and syncing using external programs.

### Features
 * Download manager: simply add songs using [youtube-dl](https://github.com/ytdl-org/youtube-dl)
 * Easily edit [ID3v2 tags](https://en.wikipedia.org/wiki/ID3) like title, artist, album, year and the cover image
 * Set up FTP clients to sync your music to all your devices
 * List and search your songs by title, artist, album or year
 * Keyboard shortcuts allow faster navigation
   - `n` for loading the page where you can add new songs
   - `/` for focusing on the search bar
   - `esc` for going to the main page
 * [Import](#Importing) songs you already have
 * Very likely works on your server, [even a Raspberry Pi](#Resources) works fine

### Screenshots

##### Add page
  ![Add Songs](.github/screenshots/shub-add.png?raw=true)
The page used for adding new songs. When a download is already running, new urls will be put in a queue. A progress bar will appear on all pages to indicate if a download is running.
  
##### Album page
  ![Album page](.github/screenshots/shub-album.png?raw=true)
Show all songs that are in an album. On this page, you can also set an album image for *all* songs in it so you don't have to set it manually for every song.
  
##### Song page
  ![Song page](.github/screenshots/shub-song.png?raw=true)
This page lets you see and edit metadata (including the cover image) that will be included in the MP3 files.

##### Song listing
  ![Listing page](.github/screenshots/shub-listing.png?raw=true)
Listings show songs sorted by some criteria, e.g. by title, artist, year or search score. 

### Installation

As a first step, you clone this repository (or download a zip file), then you open a terminal/command promt in the root directory of the repository:

```
git clone https://github.com/xarantolus/sensibleHub.git && cd sensibleHub
```

Since this is a `Go` program, you can compile it quite easily after [installing Go](https://golang.org/dl/):

```
go build -mod vendor
```

If you want to move this executable elsewhere on your system, make sure to move the following files and directories to the same location: 
 * `templates`
 * `data` (if you want to keep imported songs)
 * `assets`
 * `config.json`
 * `sensibleHub` (the executable)

#### Additional requirements
This program relies on some other programs that need to be installed and be available in your $PATH:
- [youtube-dl](https://github.com/ytdl-org/youtube-dl): Used for downloading files from [all kinds of sites](https://ytdl-org.github.io/youtube-dl/supportedsites.html). Since websites change frequently and break youtube-dl, you should update it from time to time or set up an automatic update (e.g. using a cron job)
- [FFmpeg](http://ffmpeg.org/): Used for handling the many different types of media files that are available on different websites, extracting (some) metadata during imports and transcoding mp3 files for downloads

You might be able to install them using the following command:

```
apt-get install ffmpeg youtube-dl
```

You should however check if the youtube-dl version is recent (run `youtube-dl --version`) as there are frequent changes. Alternatively, try running `youtube-dl -U` to get the newest version or check out [their releases](https://github.com/ytdl-org/youtube-dl/releases/).


### Configuration
Now you can edit `config.json` (if you want to), don't include comments (after `//`): 

```jsonc
{
    "port": 128, // HTTP server port

    // FTP settings (for sync)
    "ftp": {
        "port": 1280, // FTP port
        
        // Valid FTP username/password combinations
        "users": [
            {
                "name": "user1",
                "passwd": "user1-password"
            },
            {
                "name": "user2",
                "passwd": "user2-password"
            }
        ]
    }
}
```

Depending on your system a firewall might block some ports, so make sure to run the server with sufficient permissions (`sudo`). You might also need to mark it as executable (using `chmod +x sensibleHub`). 

After that, you are ready to start the server.

```
./sensibleHub
```

Expected output:

```
2020/05/21 19:49:29 No cleanup necessary
2020/05/21 19:49:29 [FTP] Server listening on port 1280
2020/05/21 19:49:29   Go FTP Server listening on 1280
2020/05/21 19:49:29 [Web] Server listening on port 128
```

After that, you can visit the website at `http://yourserver:128/`. You can also connect using FTP at `http://yourserver:1280/`, the account must be set in the config file.

### Importing 
This program can import songs that should be included in its library:

1. Create a directory called `import` that is at the same location as the executable.
2. Move songs in there and start the server. 
3. Songs will be imported, existing metadata embedded in files is extracted.

Please note that imports will only happen on startup, not while the software is running.

Also, a warning: any file in the `data/` and `import/` directories may be deleted by the software at any time. It happens when inconsistencies are found (e.g. a song exists in the `data/` directory on disk but isn't in the index) or a song is edited. While it doesn't delete files that are used for songs (images, audio etc.), you should make a backup anyways. As all data (except the configuration file) is stored in the `data/` directory, you can just zip it and call it a backup. 

### Resources
This program tries not to need *too much* memory or processing power, but some things are necessary. The biggest memory hog is an in-memory cache of 60x60 preview cover images.

With a library of about 750 songs and a populated cache it needs about 250Mb RAM.

I personally run it on a Raspberry Pi 4 (4GB version) and it works great. Listing pages with all songs are generated in about 300 milliseconds, but due to [InstantClick](http://instantclick.io/) it *feels* a bit faster.

### Browser support
The website should work in most modern browsers. It uses [native image lazy loading](https://caniuse.com/#feat=loading-lazy-attr) which is not yet supported by all browsers, but images will load without it regardless. If you use a recent browser version, it will be just a bit snappier. 

Everything also works *without JavaScript*, but the experience is *much better* if it's enabled ([Progressive enhancement](https://en.wikipedia.org/wiki/Progressive_enhancement)).

Mobile support also works great, menus are collapsed at the top right.

  ![Mobile listing](.github/screenshots/shub-mobile-listing.png?raw=true)

### Acknowledgements
This program would not be possible without work done by many others. For that, I would like to thank them. Here's a list of projects that are used in one way or another:

- [youtube-dl](https://github.com/ytdl-org/youtube-dl): easy tool for downloading all kinds of videos and audios
- [FFmpeg](http://ffmpeg.org/): exceptional program for handling basically [any media format](https://ffmpeg.org/ffmpeg-codecs.html) in existence
- [Go](https://golang.org/): the programming language used. It's so nice that you can have one codebase that works on so many platforms, with a very rich standard library
- [id3v2 library](https://github.com/bogem/id3v2) for editing MP3 tags
- [exiffix](https://github.com/edwvee/exiffix), [imaging](https://github.com/disintegration/imaging), [resize](https://github.com/nfnt/resize) and [goexif](https://github.com/rwcarlsen/goexif) for handling cover images *correctly*
- [FTP server library](https://goftp.io/server) for creating a virtual filesystem accessible over FTP
- [gorilla/mux](https://github.com/gorilla/mux) and [gorilla/websocket](https://github.com/gorilla/websocket) for nice HTTP server improvements, including live events over WebSockets
- [InstantClick](https://instantclick.io/): Makes the website feel significantly faster
- [Bulma](https://bulma.io): CSS framework used for designing the website
- [ReconnectingWebSocket](https://github.com/joewalnes/reconnecting-websocket): makes working with WebSockets easier

### Issues & Contributing
If you have any ideas, a pull request or something just doesn't work please feel free to get in contact.

### [License](LICENSE)
This is free as in freedom software. Do whatever you like with it.
