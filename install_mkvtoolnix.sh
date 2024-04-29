#! /usr/bin/bash
set -e

export DEBIAN_FRONTEND=noninteractive

apt-get update
apt-get install -y --no-upgrade wget
wget -O /usr/share/keyrings/gpg-pub-moritzbunkus.gpg https://mkvtoolnix.download/gpg-pub-moritzbunkus.gpg
cat <<EOM >/etc/apt/sources.list.d/mkvtoolnix.download.list
deb [signed-by=/usr/share/keyrings/gpg-pub-moritzbunkus.gpg] https://mkvtoolnix.download/debian/ bookworm main
deb-src [signed-by=/usr/share/keyrings/gpg-pub-moritzbunkus.gpg] https://mkvtoolnix.download/debian/ bookworm main
EOM
apt-get update
apt-get install -y --no-upgrade mkvtoolnix
apt-get remove -y wget
apt-get autoremove -y
apt-get clean

