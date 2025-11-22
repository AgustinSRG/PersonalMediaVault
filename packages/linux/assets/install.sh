#!/bin/bash

./uninstall.sh

cp -f ./usr/bin/pmv /usr/bin/pmv
cp -f ./usr/bin/pmvd /usr/bin/pmvd
cp -f ./usr/bin/pmv-backup /usr/bin/pmv-backup
cp -f ./usr/bin/pmv-gui /usr/bin/pmv-gui

cp -f ./usr/share/pixmaps/pmv.svg /usr/share/pixmaps/pmv.svg
cp -f ./usr/share/applications/pmv.desktop /usr/share/applications/pmv.desktop

mkdir -p /usr/lib/pmv
cp -rf ./usr/lib/pmv/www /usr/lib/pmv/www
