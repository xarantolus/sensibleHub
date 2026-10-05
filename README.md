# sensibleHub
sensibleHub is a self-hosted music management server with a web app. It lets you manage and play your music collection from any device with a web browser, and sync it to your devices using external programs.


### Features
* A single-page web app (Vue 3) that works on desktop and mobile, with automatic light/dark theme
* **Player** in the style of YouTube Music:
  * Endless play: the server keeps suggesting what to play next
  * Mini player bar while you browse, expandable to a full player with the queue
  * OS media controls (lock screen, notification shade, media keys) with cover art
  * Pop-out mini player window in browsers that support it (Document Picture-in-Picture)
  * Loudness normalisation, so all songs play equally loud
* **Next-song suggestions** based on audio analysis: key ([Camelot](https://mixedinkey.com/harmonic-mixing-guide/) harmonic mixing), tempo, energy and timbre. Newest songs are favoured and only [synced](#syncing) songs are used. Analysis is pure Go, runs on the CPU in the background and needs no GPU or external service
* **Offline playback**: upcoming songs are cached ahead, and the app and your library keep working without a connection (see [Offline and HTTPS](#offline-and-https))
* Installable as a PWA
* Easily edit [ID3v2 tags](https://en.wikipedia.org/wiki/ID3) like title, artist, album, year and the cover image, and trim the start/end of songs
* [Import](#importing) songs you already have
* Set up FTP clients to [sync](#syncing) your music to all your devices
* Download manager: simply add songs using [yt-dlp](https://github.com/yt-dlp/yt-dlp)
* Automagic metadata extraction (including cover images)
* List and search your songs by title, artist, album or year
* Live updates: changes and downloads show up on all open devices via server-sent events
* A typed [REST API](#api) with OpenAPI documentation
* Existing data is migrated automatically when you upgrade
* "Stats for nerds" setting that shows the analysis values and why a song was suggested
* Keyboard shortcuts for faster navigation
* Very likely works on your server, [even a Raspberry Pi](#resources) works fine


### Screenshots

##### Home
![Home page](.github/screenshots/home-desktop-light.png?raw=true)
The newest songs. Hover a song to play it, or use "Play all".

##### Song page with the full player
![Song page with the full player](.github/screenshots/song-player-desktop-light.png?raw=true)
The full player shows the queue ("Up next"), which you can reorder. The song page behind it lets you edit metadata and the cover image.

##### Mobile
<table>
<tr>
<td width="50%"><img src=".github/screenshots/artist-mini-player-mobile-dark.png?raw=true" alt="Artist page with the mini player"></td>
<td width="50%"><img src=".github/screenshots/player-mobile-dark.png?raw=true" alt="Full player on mobile"></td>
</tr>
</table>

An artist page with the mini player bar at the bottom (left) and the full player (right), here in dark mode.


### Installation and running
There are several methods for installing this software. Using Docker is the easiest, but you can also download release binaries or build from source. Release binaries and the Docker image contain the web app, nothing else needs to be installed for it.

After installing, look into the [configuration](#configuration) section below.

#### Docker
If you prefer using Docker, you can first run the following command to download the default configuration file:

    curl -L https://raw.githubusercontent.com/xarantolus/sensibleHub/master/config.json > config.json

Then you can already download/run the server for the first time:

    docker run -v"$(pwd):/config" -v"$(pwd)/data:/data" -p 128:128 -p 1280:1280 ghcr.io/xarantolus/sensiblehub:master

The volume mounted at `/config` must contain a `config.json` file. To change the exposed ports, you can modify the first port (before the `:`) in the command to another port. If you didn't change the configuration file `128` is the default HTTP port, `1280` is used for FTP.

You can now continue with the [configuration section](#configuration). You need to restart the container for it to use the new configuration.

#### Binaries
You can download releases from the [releases section](https://github.com/xarantolus/sensibleHub/releases/latest) of this repository.

Unzip the downloaded file to a directory of your choice on your server. Afterwards you should make sure that `yt-dlp` (or another compatible downloader) and `ffmpeg` are installed.

**Additional requirements**
This program relies on some other programs that need to be installed and be available in your $PATH:
- [yt-dlp](https://github.com/yt-dlp/yt-dlp): Used for downloading files from [all kinds of sites](https://ytdl-org.github.io/youtube-dl/supportedsites.html). Since websites change frequently and break it, you should update it from time to time or set up automatic updates (e.g. using a cron job).
- [FFmpeg](http://ffmpeg.org/) and FFprobe: Used for handling the many different types of media files that are available on different websites, extracting (some) metadata during imports, transcoding MP3 files for downloads and decoding audio for the analysis

You might be able to install them using the following command:

```
apt-get install ffmpeg python3 python3-pip && pip3 install -U yt-dlp
```

You should however check that the `yt-dlp` version is recent (run `yt-dlp --version`) as there are frequent changes. Alternatively, try running `yt-dlp --update` to get the newest version or check out [their releases](https://github.com/yt-dlp/yt-dlp/releases).

You can also put both executables in the same directory this program is installed into. That way, it should be able to find them just fine.

Depending on your system some ports might be restricted, so make sure to set sufficient permissions (see [here](https://stackoverflow.com/q/413807)). You might also need to mark the binary as executable (using `chmod +x sensibleHub`).

After that, you are ready to start the server.

```
./sensibleHub
```

Expected output:

```
2020/05/26 20:01:48 [Cleanup] No cleanup necessary
2020/05/26 20:01:48 [FTP] Server listening on port 1280
2020/05/26 20:01:48 [Web] Server listening on port 128
```

You can now continue with the [configuration section](#configuration).

#### Build from source
<details>
<summary>If no recent build is available, you can also build for yourself.</summary>

You need [Go](https://golang.org/dl/) and [Node.js 24](https://nodejs.org/) (the web app is built with Vite and embedded into the executable). Clone this repository (or download a zip file) and build with `make`:

```
git clone https://github.com/xarantolus/sensibleHub.git && cd sensibleHub
make build
```

This generates the API client, builds the web app and compiles the `sensibleHub` executable. It only redoes what changed.

If you want to move this executable elsewhere on your system, make sure to move the following files and directories to the same location:
 * `data` (if you want to keep imported songs)
 * `config.json`
 * `sensibleHub` (the executable)

To do this, you can also use the [`pack.sh`](pack.sh) script, it will create a zip file with all required assets:

```
./pack.sh
```

If you want to build for another operating system, it's quite easy. Search the correct `$GOOS` and `$GOARCH` values from [here](https://golang.org/doc/install/source#environment) and add them to the command. For the Raspberry Pi, the following values can be used:

```
GOOS=linux GOARCH=arm GOARM=7 ./pack.sh
```

</details>


### Configuration

<details><summary>You can now edit <code>config.json</code>, open this to get more info.</summary>

The following file details the configuration options. You can use comments (`//`) in the configuration file.

```jsonc
{
    // HTTP server port (used for accessing the website)
    "port": 128,

    // FTP settings
    "ftp": {
        // FTP port the server will listen on. You will need this when setting up syncing
        "port": 1280,

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
    },

    // How long generated files are kept, in days.
    // A small number means that less storage is used in general, but files will be generated with every sync/download (if there are changes).
    // If negative, they will be kept forever, if zero they will not be kept.
    // Files are checked every day at 0:00.
    // If you use multiple devices that sync at different intervals, it is recommended to keep files for a few days.
    "keep_generated_days": 3,

    // External data sources can be disabled
    "allow_external": {
        // If set to true, a search query to iTunes will be sent to get a high-quality cover image when downloading a new song.
        "apple": true
    },

    // Settings for cover images. Affects only those in generated MP3 files
    "cover": {
        // Cover images of generated/synced songs will have this as maximum size in pixels, larger ones are downscaled.
        // If omitted, 0 or lower, this setting will be ignored and image sizes are not changed.
        "max_size": 2000
    },

    // Alternatives for programs used by this server. Leave blank to use default values.
    // Allows you to set alternative paths for programs, e.g. if you want to use an alternative youtube-dl fork such as [this one](https://github.com/yt-dlp/yt-dlp)
    // This section is ignored if running in Docker
    "alternatives": {
        "ffmpeg": "ffmpeg",
        "ffprobe": "ffprobe",
        "youtube-dl": "yt-dlp"
    },

    // Whether to generate cover previews when starting up.
    // If this is false, cover previews are first generated the first time a page is loaded, which
    // can lead to pages where previews come in after serveral seconds
    "generate_on_startup": true,

    // Songs are analysed in the background (key, tempo, loudness, timbre) to pick fitting songs
    // to play next and to play every song equally loud. Requires ffmpeg. Set to true to turn it off.
    // Optional: if omitted, analysis is enabled.
    "analysis": {
        "disabled": false
    }
}
```

With `analysis.disabled` set to `true`, no song is analysed. Next songs are then picked randomly and no loudness normalisation is applied.

</details>


Assuming you kept the default ports, you can visit the website at `http://yourserver:128/`. You can also connect via FTP at `ftp://yourserver:1280/` using one of the accounts set in the config file.


### Importing
This program can import songs that should be included in its library in a few different ways.

##### From disk (usually for initially importing your library)

1. Create a directory called `import` that is at the same location as the executable.
2. **Copy** songs into the `import` directory. It does not matter if you copy the files directly or directories containing them (the server will search everything in there). Please note that **the server will delete files from the import directory** once they are added to its library.
3. (Re)start the server.
4. Songs will be imported, existing metadata embedded in files is extracted.

These imports will only happen on startup, not while the software is running, so you have to restart the server every time with this method.

##### Over network/FTP
You can also import files by putting them in *any* directory over FTP. On Windows, you can [create a FTP network connection](https://superuser.com/a/88572) quite nicely.

Now any music file that is moved there will be imported. It seems like import errors are **not** shown, so you might need to watch the server output to see if anything went wrong.

Also, a warning: any file in the `data/` and `import/` directories may be deleted by the software at any time. It happens when inconsistencies are found (e.g. a song exists in the `data/` directory on disk but isn't in the index) or a song is edited. While it doesn't delete files that are used for songs (images, audio etc.), you should make a backup anyways. As all data (except the configuration file) is stored in the `data/` directory, you can just zip it and call it a backup.


### Syncing
Obviously one wants to have their music with them on all devices, even when offline. Here's a guide on how to achieve that on Windows/Linux desktop and Android.


##### Desktop
On a PC or Laptop, you can create recurring sync jobs (on all platforms) that use [rclone](https://github.com/rclone/rclone) (which you need to install before continuing).

First, [set up a new rclone FTP remote](https://rclone.org/ftp/) with `rclone config`. After setup, it should look similar to this:

```
[MyMusic]
type = ftp
host = yourserver
user = myusername
port = 1280
pass = *** ENCRYPTED ***
```

Now you can use rclone sync like this to sync it to your music directory:

```
rclone sync --update --ignore-size -v MyMusic:/ %USERPROFILE%\Music
```

The `--ignore-size` flag is very important as the server doesn't always know the correct file size if the file hasn't been generated yet.

If you want to, you can set this up as a cron job or use windows task scheduler to run the command automatically. Another simple option is creating a batch file/script and running it from time to time.

My recommended music player for Windows is [Dopamine](https://github.com/digimezzo/dopamine-windows), it can automatically index the music directory. You can [download it here](https://www.digimezzo.com/content/software/dopamine/).


##### Android
On Android, you can use any FTP app that doesn't look at the file size or lets you disable that. One of them is [FolderSync](https://play.google.com/store/apps/details?id=dk.tacit.android.foldersync.lite).

Add a new "account" (in-app, there's no registration) with the following attributes:
 * **Server address**: the server name, e.g. `yourserver`.
 * **Port**: the FTP port you set in the configuration file, e.g. `1280`
 * **Login name/password**: Your login credentials from one of the FTP users set in the [config file](#Configuration)
 * The path can be left empty

Now you can create a new *Folder pair* with these settings:
 * **Account**: The one created above
 * **Sync type**: to local folder
 * **Remote folder**: should be empty or just `/`
 * **Local Folder**: Your Android music folder, might be `/storage/emulated/0/Music`
 * **Scheduling**: Here you can set *when* it should sync your files
 * **Sync options**: Enable *Sync subfolders* and *Sync deletions*.
 * **Advanced settings**
   * **Overwrite old files**: always
   * **If both local and remote file are modified**: *Use remote file*
   * **Rescan media library** should be on, that way new files are imported
   * **Disable file-size check** should be on, **this is the most important setting**

For Android, any music player will probably work. I recommend [Music](https://f-droid.org/packages/com.maxfour.music/), it is quite customizable and colorful. You can enable *Ignore Media Store covers* in settings if some cover images aren't displayed.


### Resources
This program tries not to need *too much* memory.

I personally run it on a Raspberry Pi 4 (4GB version) and it works great. The whole song list is loaded once and cached in the browser, so browsing is instant afterwards. Audio analysis takes a few seconds per song and runs in the background, one song at a time, so the server stays responsive while it catches up on a big library. You can turn it off with `analysis.disabled`.

RAM usage is a bit weird. While on windows (where I develop) everything seems to be around 50MB, it looks like there's a problem on ARM computers (like the Raspberry Pi):
using the same music library it needs about ten times as much memory. I have *not* found out where this issue comes from.


### Assumptions
There are several assumptions made so the program will work as expected in most cases.

- Two songs have the same artist if the artist attribute is not empty and equal (case insensitive) after being put through the `CleanName` function in [`store/album.go`](store/album.go).
- Two songs are in the same album if that attribute is not empty, the above applies for the artist and the same applies for the album name.
- A song should have *one* artist, every other performer is mentioned in brackets in the song title, e.g. like `Title (feat. Artist2 & Artist3)`. If this is not done, the "Featured in" listing of the artists' page might not display all relevant songs.
- All cover images are squared. Any that aren't will be cropped and some part of the image will be removed.


### Keyboard shortcuts
These are keyboard shortcuts that can be used on any page:
- `n` for loading the page where you can add new songs
- Listings: `s` for all songs, `a` for artists, `y` for years, `i` for incomplete, `e` for recent edits and `u` for unsynced songs
- `/` for focusing on the search bar
- `esc` for going to the main page
- `space` for play/pause, `shift` + `←`/`→` for previous/next song


### Offline and HTTPS
Offline playback and installing the app as a PWA rely on a service worker and Cache Storage, which browsers only allow on secure origins: **HTTPS or `localhost`**.

Over plain `http://` on a LAN address (e.g. `http://192.168.1.10:128`) the app works as usual, but without the offline features: nothing is cached ahead and a reload without a connection fails. sensibleHub does not serve HTTPS itself, so to get the offline features on other devices, put it behind a reverse proxy that terminates TLS (e.g. [Caddy](https://caddyserver.com/), nginx or Traefik) and open the app through that address.


### Browser support
The web app needs a recent browser. Offline features need a [secure origin](#offline-and-https). The pop-out player needs Document Picture-in-Picture (Chromium-based desktop browsers), and OS media controls need the Media Session API; where a browser lacks an API, that feature is simply not offered. The layout adapts to mobile screens.


### Development
You need Go, Node 24 and `make`. Everything goes through the [`Makefile`](Makefile):

```
make gen      # Go API structs -> frontend/openapi.json -> frontend/src/api/schema.ts
make check    # gofmt, go vet, go test, then vue-tsc, eslint and vitest
make build    # gen, build the frontend, compile the binary (only redoes what changed)
```

`make gen-check` regenerates the API files and fails if the committed ones differ; CI runs it. The generated `frontend/openapi.json` and `frontend/src/api/schema.ts` are committed.

The binary embeds `frontend/dist`, so Go code only compiles after the frontend has been built once (`make build` or `make frontend-build`).

For working on the frontend with hot reload, run the Go server and the Vite dev server side by side. `SH_BACKEND` is the address Vite proxies `/api` and `/media` to (default `http://localhost:128`):

```
go run -mod vendor . -debug
cd frontend && SH_BACKEND=http://localhost:<port> npm run dev
```

`<port>` is the `port` from your `config.json`. `ffmpeg`, `ffprobe` and `yt-dlp` must be on your `PATH` when running the server, but not for building.

**Where the API is defined:** Go structs in [`web/api`](web/api) (mainly `dto.go`) are the single source of truth. They produce the OpenAPI spec (`cmd/openapi` prints it without starting a server), and `openapi-typescript` turns that into the typed TypeScript client used by the frontend. Don't write request/response types by hand in the frontend.


### API
The web app uses a typed REST API under `/api/v1`, which you can use too. Errors are returned as RFC 9457 `application/problem+json`, live updates are a server-sent events stream at `/api/v1/events`.

* Interactive documentation: `/api/v1/docs`
* OpenAPI spec: `/api/v1/openapi.json`

Media files (covers, audio, generated MP3s) are served under `/media/songs/{id}/...`.


### Limitations
Compared to other music servers this one is very basic. Here are some things you should be aware of:

* It does **not** support the [SubSonic API](http://www.subsonic.org/pages/api.jsp). You can not use this software as a back-end for SubSonic-compatible music players.
* Some **metadata will be lost** when importing: everything except for the cover image, title, artist, album and year will be **discarded**. Keep a backup of your music before importing.
* Does not serve HTTPS itself and has no web login (only the FTP server has accounts). The software is intended to be hosted inside a local network *only*; use a reverse proxy for [HTTPS](#offline-and-https).
* Songs in albums are not sorted by their title numbers, but alphabetically. If there's a song with the same title as the album itself, it will be the first song.
* The song list is not paginated: the web app loads all songs at once and derives every view from it. With a very large music collection this might become slow, depending on your device.
* Song suggestions only consider songs that are marked for syncing, and need the background analysis to have finished for a song to be matched by its audio. Until then, songs are picked randomly.
* As song IDs use 52 characters and have a length of 4, you are limited to 52^4 = 7.311.616 songs. The server might crash when generating a new ID before you reach that limit (when it doesn't find an unused ID the first 10.000 times).
* It seems like some media players don't display cover images over a certain size, while others do. Use the cover max size setting to see if lowering the size helps. On Android, use a music player that allows you to ignore MediaStore covers.

### Acknowledgements
This program would not be possible without work done by many others. For that, I would like to thank them. Here's a list of projects that are used in one way or another:

- [yt-dlp](https://github.com/yt-dlp/yt-dlp): easy tool for downloading all kinds of videos and audios
- [FFmpeg](http://ffmpeg.org/): exceptional program for handling basically [any media format](https://ffmpeg.org/ffmpeg-codecs.html) in existence
- [Go](https://golang.org/): the programming language used. It's so nice that you can have one codebase that works on so many platforms, with a very rich standard library
- [id3v2 library](https://github.com/bogem/id3v2) for reading MP3 tags
- [exiffix](https://github.com/edwvee/exiffix), [imaging](https://github.com/disintegration/imaging), [resize](https://github.com/nfnt/resize) and [goexif](https://github.com/rwcarlsen/goexif) for handling cover images *correctly*
- [FTP server library](https://goftp.io/server) for creating a virtual filesystem accessible over FTP
- [gorilla/mux](https://github.com/gorilla/mux) for the HTTP router and [Huma](https://huma.rocks/) for the typed API, OpenAPI spec and server-sent events
- [gonum](https://www.gonum.org/) for the FFT used in the audio analysis
- [Vue](https://vuejs.org/), [Vite](https://vite.dev/), [TanStack Query](https://tanstack.com/query), [Workbox](https://developer.chrome.com/docs/workbox) and [openapi-typescript](https://openapi-ts.dev/) for the web app
- [Bulma](https://bulma.io) and [Oruga](https://oruga-ui.com/): CSS framework and component library used for designing the website


### Issues & Contributing
If you have any ideas, a pull request or something just doesn't work please feel free to get in contact.


### [License](LICENSE)
This is free as in freedom software. Do whatever you like with it.
