#!/bin/bash

cd ppa

# Create Key.gpg

gpg --armor --export "${EMAIL}" > KEY.gpg

# Create Packages and Packages.gz

dpkg-scanpackages --multiversion . > Packages

gzip -k -f Packages

# Create Release file

apt-ftparchive release . > Release

# Create signed release files

gpg --default-key "${EMAIL}" -abs -o - Release > Release.gpg
gpg --default-key "${EMAIL}" --clearsign -o - Release > InRelease

# Create list file

echo "deb [signed-by=/etc/apt/trusted.gpg.d/pmv.gpg] https://${GITHUB_USERNAME}.github.io/PersonalMediaVault/ ./" > pmv.list
