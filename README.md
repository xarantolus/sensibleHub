# sensibleHub
sensibleHub is a self-hosted music management server. It allows managing your music collection from any device (that has a web browser) 
and syncing using external programs.

### Features
 * Download manager: simply add songs using [youtube-dl](https://github.com/ytdl-org/youtube-dl)
 * Easily edit [ID3v2 tags](https://en.wikipedia.org/wiki/ID3) like title, artist, album, year and the cover image
 * Set up FTP clients to [sync](#Syncing) your music to all your devices
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
  ![Album page](.github/screenshots/shub-album.jpg?raw=true)
Show all songs that are in an album. On this page, you can also set an album image for *all* songs in it so you don't have to set it manually for every song.
  
##### Song page
  ![Song page](.github/screenshots/shub-song.png?raw=true)
This page lets you see and edit metadata (including the cover image) that will be included in the MP3 files. 

##### Song listing
  ![Listing page](.github/screenshots/shub-listing.png?raw=true)
Listings show songs sorted by some criteria, e.g. by title, artist, year or search score. 

##### Additional listings
In the "More" menu at the upper right side, you can find other listings that can be useful for metadata editing.

<p align="center">
<img src=".github/screenshots/shub-additional-listings.png?raw=true" width="50%">
</p>

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

To do this, you can also use the [`pack.sh`](pack.sh) script, it will create a zip file with all required assets:

```
./pack.sh
```

If you want to build for another operating system, it's quite easy. Search the correct `$GOOS` and `$GOARCH` values from [here](https://golang.org/doc/install/source#environment) and add them to the command. For the Raspberry Pi, the following values can be used:

```
GOOS=linux GOARCH=arm GOARM=7 ./pack.sh
```

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
    },

    // How long generated files are kept, in days.
    // A small number means that less storage is used in general, but files will be generated with every sync/download (if there are changes).
    // If negative, they will be kept forever, if zero they will not be kept.
    // Files are checked every day at 0:00.
    // If you use multiple devices that sync at different intervals, it is recommended to keep files for a few days.
    "keep_generated_days": 3
}
```

Depending on your system a firewall might block some ports, so make sure to run the server with sufficient permissions (`sudo`). You might also need to mark it as executable (using `chmod +x sensibleHub`). 

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

After that, you can visit the website at `http://yourserver:128/`.
You can also connect via FTP at `ftp://yourserver:1280/` using one of the accounts set in the config file.

### Importing 
This program can import songs that should be included in its library in a few different ways. 

##### From disk

1. Create a directory called `import` that is at the same location as the executable.
2. Move songs in there and start the server. 
3. Songs will be imported, existing metadata embedded in files is extracted.

Please note that these imports will only happen on startup, not while the software is running.

##### Over network/FTP
You can also import files by putting them in *any* directory over FTP. On windows, you can [create a FTP network connection](https://superuser.com/a/88572) quite nicely.

Now any music file that is moved there will be imported. It seems like import errors are **not** shown, so you might need to watch the log file if anything went wrong.

Also, a warning: any file in the `data/` and `import/` directories may be deleted by the software at any time. It happens when inconsistencies are found (e.g. a song exists in the `data/` directory on disk but isn't in the index) or a song is edited. While it doesn't delete files that are used for songs (images, audio etc.), you should make a backup anyways. As all data (except the configuration file) is stored in the `data/` directory, you can just zip it and call it a backup. 

### Syncing
Obviously one wants to have their music with them on all devices, even when they are offline. Here's a guide on how to achieve that.

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

If you want to, you can set this up as a cron job or use windows task scheduler to run the command automatically.

##### Android
On Android, you can use any FTP app that doesn't look at the file size or lets you disable that. One of them is [FolderSync](https://play.google.com/store/apps/details?id=dk.tacit.android.foldersync.lite).

Add a new "account" (in-app, there's no registration) with the following attributes:
 * **Server address**: `ftp://yourserver:1280`
 * **Login**: Your login credentials from one of the FTP users set in the [config file](#Configuration)

Now you can create a new *Folder pair* with these settings:
 * **Account**: The one created above
 * **Local Folder**: Your music folder, might be `/storage/emulated/0/Music`
 * **Scheduling**: Here you can set *when* it should sync your files
 * **Sync options**: Enable *Sync subfolders* and *Sync deletions*. Disable *Only resync source files if modified (ignore target deletion)*. Set *Overwrite old files* to *Always* and set *If conflicting modifications* to *Use remote file*. Now the most important part, **you must enable _Disable file-size check_** or it will not work. You can also enable *Rescan media library* to make sure new files are recognized.

### Resources
This program tries not to need *too much* memory or processing power, but some things are necessary. The biggest memory hog is an in-memory cache of 60x60 preview cover images.

With a library of about 750 songs and a populated cache it needs about 250Mb RAM.

I personally run it on a Raspberry Pi 4 (4GB version) and it works great. Listing pages with all songs are generated in about 300 milliseconds, but due to [InstantClick](http://instantclick.io/) it *feels* a bit faster.

### Browser support
The website should work in most modern browsers. It uses [native image lazy loading](https://caniuse.com/#feat=loading-lazy-attr) which is not yet supported by all browsers, but images will load without it regardless. If you use a recent browser version, it will be just a bit snappier. 

Everything also works *without JavaScript*, but the experience is *much better* if it's enabled ([Progressive enhancement](https://en.wikipedia.org/wiki/Progressive_enhancement)).

Mobile support also works great, menus are collapsed at the top right.

Song listing             |  Song listing with opened menu
:-------------------------:|:-------------------------:
![Mobile listing](.github/screenshots/shub-mobile-listing.png?raw=true) | ![Mobile Menu](.github/screenshots/shub-mobile-menu.png?raw=true)

  

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
