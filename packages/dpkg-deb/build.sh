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

PMV_VERSION_MAJOR=4
PMV_VERSION_MINOR=0
PMV_VERSION_REVISION=0

PMV_BIN_ARCH=amd64

PMV_MAINTAINER="AgustinSRG <agustinsanromanguzman@gmail.com>"
PMV_DESCRIPTION="PersonalMediaVault - Web application to store media files in an encrypted storage, and to visualize them using a web browser."

PMV_DEB_NAME=personalmediavault_${PMV_VERSION_MAJOR}.${PMV_VERSION_MINOR}.${PMV_VERSION_REVISION}_amd64

PMV_PKG_FOLDER=./${PMV_DEB_NAME}
PMV_DEB_FILE=./${PMV_DEB_NAME}.deb

echo "Preparing folder:" ${PMV_PKG_FOLDER}

rm -rf ${PMV_PKG_FOLDER}
rm -rf ${PMV_DEB_FILE}
mkdir -p ${PMV_PKG_FOLDER}

echo "Copying files..."

# Binaries + frontend

mkdir -p ${PMV_PKG_FOLDER}/usr/bin

cp ../../backend/pmvd ${PMV_PKG_FOLDER}/usr/bin/pmvd
cp ../../backup-tool/pmv-backup ${PMV_PKG_FOLDER}/usr/bin/pmv-backup
cp ../../launcher/pmv ${PMV_PKG_FOLDER}/usr/bin/pmv

mkdir -p ${PMV_PKG_FOLDER}/usr/lib/pmv

cp -rf ../../frontend/dist ${PMV_PKG_FOLDER}/usr/lib/pmv/www

# Application icon

mkdir -p ${PMV_PKG_FOLDER}/usr/share/pixmaps/
cp ./assets/pmv.svg ${PMV_PKG_FOLDER}/usr/share/pixmaps/pmv.svg

# Application desktop entry

mkdir -p ${PMV_PKG_FOLDER}/usr/share/applications
cp ./assets/pmv.desktop ${PMV_PKG_FOLDER}/usr/share/applications/pmv.desktop

# Custom actions for Nemo file explorer

mkdir -p ${PMV_PKG_FOLDER}/usr/share/nemo/actions
cp ./assets/pmv.nemo_action ${PMV_PKG_FOLDER}/usr/share/nemo/actions/pmv.nemo_action
cp ./assets/pmv-noselect.nemo_action ${PMV_PKG_FOLDER}/usr/share/nemo/actions/pmv-noselect.nemo_action

echo "Configuring package..."

mkdir -p ${PMV_PKG_FOLDER}/DEBIAN

CONTROL_FILE=${PMV_PKG_FOLDER}/DEBIAN/control

echo "Package: personalmediavault" > ${CONTROL_FILE}
echo "Version:" ${PMV_VERSION_MAJOR}.${PMV_VERSION_MINOR}-${PMV_VERSION_REVISION} >> ${CONTROL_FILE}
echo "Section: web" >> ${CONTROL_FILE}
echo "Priority: optional" >> ${CONTROL_FILE}
echo "Architecture:" ${PMV_BIN_ARCH} >> ${CONTROL_FILE}
echo "Depends: libc6,ffmpeg" >> ${CONTROL_FILE}
echo "Maintainer:" ${PMV_MAINTAINER} >> ${CONTROL_FILE}
echo "Description:" ${PMV_DESCRIPTION} >> ${CONTROL_FILE}

echo "Building package..."

dpkg-deb --build ${PMV_PKG_FOLDER}

echo "Cleaning up..."

rm -r ${PMV_PKG_FOLDER}

if [ "$1" = "ppa" ]
then
    mkdir -p ./ppa
    mv ${PMV_DEB_FILE} ./ppa/${PMV_DEB_NAME}.deb
fi

echo "DONE!"


