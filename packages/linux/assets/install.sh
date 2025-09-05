#!/bin/bash

./uninstall.sh

cp -f ./usr/bin/pmv /usr/bin/pmv
cp -f ./usr/bin/pmvd /usr/bin/pmvd
cp -f ./usr/bin/pmv-backup /usr/bin/pmv-backup

mkdir -p /usr/lib/pmv
cp -rf ./usr/lib/pmv/www /usr/lib/pmv/www
