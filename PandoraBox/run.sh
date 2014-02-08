#! /bin/bash

pkill vlc #For saftefy
vlc -d -I http --http-port 8000
/home/ubuntu/gocode/bin/PandoraBox &> /home/ubuntu/boxClient.log
pkill vlc #Clean up after yourself