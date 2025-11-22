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

cd ../launcher-gui
echo "Building GUI launcher..."
./build-production.sh

cd ../frontend
echo "Building frontend..."
npm install
npm run build

cd ../packages/linux

# Build package

PMV_VERSION_MAJOR=4
PMV_VERSION_MINOR=0
PMV_VERSION_REVISION=1

PMV_BIN_ARCH=amd64

PMV_PKG_NAME=personalmediavault_${PMV_VERSION_MAJOR}.${PMV_VERSION_MINOR}.${PMV_VERSION_REVISION}_amd64

PMV_PKG_FOLDER=./${PMV_PKG_NAME}
PMV_TAR_NAME=./${PMV_PKG_NAME}.tar.gz

echo "Preparing folder:" ${PMV_PKG_FOLDER}

rm -rf ${PMV_PKG_FOLDER}
rm -rf ${PMV_TAR_NAME}
mkdir -p ${PMV_PKG_FOLDER}

echo "Copying files..."

# Binaries + frontend

mkdir -p ${PMV_PKG_FOLDER}/usr/bin

cp ../../backend/pmvd ${PMV_PKG_FOLDER}/usr/bin/pmvd
cp ../../backup-tool/pmv-backup ${PMV_PKG_FOLDER}/usr/bin/pmv-backup
cp ../../launcher/pmv ${PMV_PKG_FOLDER}/usr/bin/pmv
cp ../../launcher-gui/target/release/pmv-gui ${PMV_PKG_FOLDER}/usr/bin/pmv-gui

mkdir -p ${PMV_PKG_FOLDER}/usr/lib/pmv

cp -rf ../../frontend/dist ${PMV_PKG_FOLDER}/usr/lib/pmv/www

# Desktop entry

mkdir -p ${PMV_PKG_FOLDER}/usr/share/applications
mkdir -p ${PMV_PKG_FOLDER}/usr/share/pixmaps

cp ./assets/pmv.svg ${PMV_PKG_FOLDER}/usr/share/pixmaps/pmv.svg
cp ./assets/pmv.desktop ${PMV_PKG_FOLDER}/usr/share/applications/pmv.desktop

# Scripts

cp ./assets/install.sh ${PMV_PKG_FOLDER}/install.sh
cp ./assets/uninstall.sh ${PMV_PKG_FOLDER}/uninstall.sh

echo "Built folder: ${PMV_PKG_FOLDER}"

echo "Building compressed package..."

tar czf ${PMV_TAR_NAME} ${PMV_PKG_FOLDER}

echo "Built compressed package: ${PMV_TAR_NAME}"

echo "DONE!"


