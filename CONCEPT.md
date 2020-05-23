# sensibleHub
Music Management Server


# Features
 - Serviceunabhängiges Abspeichern von Musikdateien (youtube-dl)
 - Erhalten der Original-Datei
 - On-fly generierung einer verwendbaren MP3-Datei mit Cover, Tags...
 - Speicherung eines Covers
 - Bearbeitung der Tags
 - Synchronisierung unabhängig von Synology NAS 


# Storage
Abspeichern unter `data` directory:

- data
---- all.json
---- songs
-------- (id here)
------------ cover.??? <!-- Max Size. Maybe set a specified format? -->
------------ original.???  <!-- Original sound file from youtube-dl -->
------------ latest.mp3  <!-- Maybe don't save this at all? -->