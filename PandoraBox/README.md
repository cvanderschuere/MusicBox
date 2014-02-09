Pandora Box
===========


##Dependencies

```
	#VLC (headless)
	sudo apt-get install vlc-nox

```


##Autostart

Add this file to /etc/init

Ex: /etc/init/pandorabox.conf

```
#  Pandora Box

description "Pandora Box"
version "1.0"

start on (local-filesystems and net-device-up IFACE!=lo) 
stop on runlevel [06]

respawn

setuid ubuntu

exec /home/ubuntu/gocode/src/MusicBox/PandoraBox/run.sh

```
