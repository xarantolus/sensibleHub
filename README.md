# sensibleHub
sensibleHub is a self-hosted music management server. It allows managing your music collection from any device (that has a web browser)
and syncing using external programs.


### Features
* Easily edit [ID3v2 tags](https://en.wikipedia.org/wiki/ID3) like title, artist, album, year and the cover image
* [Import](#Importing) songs you already have
* Set up FTP clients to [sync](#Syncing) your music to all your devices
* Download manager: simply add songs using [youtube-dl](https://github.com/ytdl-org/youtube-dl)
* Automagic metadata extraction (including cover images)
* List and search your songs by title, artist, album or year
* Very likely works on your server, [even a Raspberry Pi](#Resources) works fine
* Automatic dark mode: the site applies a light or dark theme depending on your system settings
* Keyboard shortcuts for faster navigation
* No JavaScript required


### Screenshots

##### Song page
  ![Song page](.github/screenshots/shub-song.png?raw=true)
This page lets you see and edit metadata, including the cover image, that will be included in the generated MP3 file. The image shows both the dark and light mode.

##### Add page
  ![Add Songs](.github/screenshots/shub-add.png?raw=true)
The page used for adding new songs. When a download is already running, new urls will be put in a queue. A progress bar will appear on all pages to indicate if a download is running.

##### Album page
  ![Album page](.github/screenshots/shub-album.png?raw=true)
Show all songs that are in an album. On this page, you can also set an album image for *all* songs in it so you don't have to set it manually for every song.

##### Song listing
  ![Listing page](.github/screenshots/shub-listing.png?raw=true)
Listings show songs sorted by some criteria, e.g. by title, artist, year or search score.

##### Additional listings
In the "More" menu at the upper right side, you can find other listings that can be useful for metadata editing.

<p align="center">
<img src=".github/screenshots/shub-additional-listings.png?raw=true" width="50%">
</p>

##### Search suggestions
While typing in the search box, your collection is already searched and suggestions are shown:

<p align="center">
<img src=".github/screenshots/suggestions.gif?raw=true" >
</p>


### Installation
There are several methods for installing this software. Using Docker is the easiest, but you can also download release binaries or build from source.

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

Unzip the downloaded file to a directory of your choice on your server. Afterwards you should make sure that `youtube-dl` (or another compatible downloader) and `ffmpeg` are installed.

**Additional requirements**
This program relies on some other programs that need to be installed and be available in your $PATH:
- [yt-dlp](https://github.com/yt-dlp/yt-dlp): Used for downloading files from [all kinds of sites](https://ytdl-org.github.io/youtube-dl/supportedsites.html). Since websites change frequently and break it, you should update it from time to time or set up automatic updates (e.g. using a cron job).
- [FFmpeg](http://ffmpeg.org/) and FFprobe: Used for handling the many different types of media files that are available on different websites, extracting (some) metadata during imports and transcoding MP3 files for downloads

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

As a first step, you clone this repository (or download a zip file), then you open a terminal/command prompt in the root directory of the repository:

```
git clone https://github.com/xarantolus/sensibleHub.git && cd sensibleHub
```

Since this is a `Go` program, you can compile it quite easily after [installing Go](https://golang.org/dl/):

```
go build -mod vendor
```

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
    "generate_on_startup": true
}
```

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

I personally run it on a Raspberry Pi 4 (4GB version) and it works great. Listing pages with all songs are generated in about 300 milliseconds, but due to [InstantClick](http://instantclick.io/) it *feels* a bit faster.

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


### Browser support
The website should work in most modern browsers. It uses [native image lazy loading](https://caniuse.com/#feat=loading-lazy-attr) which is not yet supported by all browsers, but images will load without it regardless. If you use a recent browser version, it will be just a bit snappier.

Everything also works *without JavaScript*, but the experience is *much better* if it's enabled ([Progressive enhancement](https://en.wikipedia.org/wiki/Progressive_enhancement)).

Mobile support also works great, menus are collapsed at the top right.

|                              Song listing                               |                   Song listing with opened menu                   |
| :---------------------------------------------------------------------: | :---------------------------------------------------------------: |
| ![Mobile listing](.github/screenshots/shub-mobile-listing.png?raw=true) | ![Mobile Menu](.github/screenshots/shub-mobile-menu.png?raw=true) |


### Limitations
Compared to other music servers this one is very basic. Here are some things you should be aware of:

* It does **not** support the [SubSonic API](http://www.subsonic.org/pages/api.jsp). You can not use this software as a back-end for SubSonic-compatible music players.
* Some **metadata will be lost** when importing: everything except for the cover image, title, artist, album and year will be **discarded**. Keep a backup of your music before importing.
* Does not support HTTPS. The software is intended to be hosted inside a local network *only*.
* Songs in albums are not sorted by their title numbers, but alphabetically. If there's a song with the same title as the album itself, it will be the first song.
* The web interface does not split long lists into multiple pages. If you have a large music collection, loading a page might be limited by your browsers' performance (the server should be able to generate the necessary HTML just fine, but then generating cover previews might become a problem). My guess is that this will happen, depending on your device, at about 10.000 songs.
* As song IDs use 52 characters and have a length of 4, you are limited to 52^4 = 7.311.616 songs. The server might crash when generating a new ID before you reach that limit (when it doesn't find an unused ID the first 10.000 times).
* It seems like some media players don't display cover images over a certain size, while others do. Use the cover max size setting to see if lowering the size helps. On Android, use a music player that allows you to ignore MediaStore covers.

### Acknowledgements
This program would not be possible without work done by many others. For that, I would like to thank them. Here's a list of projects that are used in one way or another:

- [youtube-dl](https://github.com/ytdl-org/youtube-dl): easy tool for downloading all kinds of videos and audios
- [FFmpeg](http://ffmpeg.org/): exceptional program for handling basically [any media format](https://ffmpeg.org/ffmpeg-codecs.html) in existence
- [Go](https://golang.org/): the programming language used. It's so nice that you can have one codebase that works on so many platforms, with a very rich standard library
- [id3v2 library](https://github.com/bogem/id3v2) for reading MP3 tags
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
