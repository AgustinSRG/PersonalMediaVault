#!/bin/bash

# Build project

echo "Building project: PersonalMediaVault"

cd ../../backend
echo "Building backend..."
./build-production.sh

cd ../backup-tool
echo "Building backup tool..."
./build-production.sh

cd ../launcher
echo "Building launcher..."
./build-production.sh

cd ../frontend
echo "Building frontend..."
npm install
npm run build

cd ../packages/dpkg-deb

# Build package

PMV_VERSION_MAJOR=1
PMV_VERSION_MINOR=14
PMV_VERSION_REVISION=0

PMV_BIN_ARCH=amd64

PMV_MAINTAINER="AgustinSRG <agustinsanromanguzman@gmail.com>"
PMV_DESCRIPTION="PersonalMediaVault - Web application to store media files in an encrypted storage, and to visualize them using a web browser."

PMV_PKG_FOLDER=./personalmediavault_${PMV_VERSION_MAJOR}.${PMV_VERSION_MINOR}-${PMV_VERSION_REVISION}

echo "Preparing folder:" ${PMV_PKG_FOLDER}

rm -rf ${PMV_PKG_FOLDER}
rm -rf ${PMV_PKG_FOLDER}.deb
mkdir -p ${PMV_PKG_FOLDER}

mkdir -p ${PMV_PKG_FOLDER}/usr/bin
mkdir -p ${PMV_PKG_FOLDER}/usr/lib/pmv
mkdir -p ${PMV_PKG_FOLDER}/DEBIAN

echo "Copying files..."

cp ../../backend/pmvd ${PMV_PKG_FOLDER}/usr/bin/pmvd
cp ../../backup-tool/pmv-backup ${PMV_PKG_FOLDER}/usr/bin/pmv-backup
cp ../../launcher/pmv ${PMV_PKG_FOLDER}/usr/bin/pmv

cp -rf ../../frontend/dist ${PMV_PKG_FOLDER}/usr/lib/pmv/www

echo "Configuring package..."

CONTROL_FILE=${PMV_PKG_FOLDER}/DEBIAN/control

echo "Package: personalmediavault" > ${CONTROL_FILE}
echo "Version:" ${PMV_VERSION_MAJOR}.${PMV_VERSION_MINOR}-${PMV_VERSION_REVISION} >> ${CONTROL_FILE}
echo "Section: web" >> ${CONTROL_FILE}
echo "Priority: optional" >> ${CONTROL_FILE}
echo "Architecture:" ${PMV_BIN_ARCH} >> ${CONTROL_FILE}
echo "Depends: ffmpeg" >> ${CONTROL_FILE}
echo "Maintainer:" ${PMV_MAINTAINER} >> ${CONTROL_FILE}
echo "Description:" ${PMV_DESCRIPTION} >> ${CONTROL_FILE}

echo "Building package..."

dpkg-deb --build ${PMV_PKG_FOLDER}

echo "DONE!"


