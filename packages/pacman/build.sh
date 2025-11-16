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
PMV_VERSION_REVISION=0

PMV_BIN_ARCH=amd64

PMV_PKG_NAME=personalmediavault-${PMV_VERSION_MAJOR}.${PMV_VERSION_MINOR}.${PMV_VERSION_REVISION}

PMV_PKG_FOLDER=${PMV_PKG_NAME}
PMV_TAR_NAME=${PMV_PKG_NAME}.tar.gz

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

echo "Built folder: ${PMV_PKG_FOLDER}"

echo "Building compressed package..."

tar czf ${PMV_TAR_NAME} ${PMV_PKG_FOLDER}

echo "Built compressed package: ${PMV_TAR_NAME}"

PKG_BUILD_FILE=PKGBUILD

echo "Generating ${PKG_BUILD_FILE} file..."

echo "# Maintainer: Agustin San Roman <agustinsrg@air-institute.com>" > ${PKG_BUILD_FILE}

echo "" >> ${PKG_BUILD_FILE}

echo "pkgname=personalmediavault" >> ${PKG_BUILD_FILE}
echo "pkgver=2.7.1" >> ${PKG_BUILD_FILE}
echo "pkgrel=1" >> ${PKG_BUILD_FILE}
echo 'pkgdesc="Self-hosted web application to store media files (video, audio and pictures) in an encrypted storage, and visualize them using a web browser."' >> ${PKG_BUILD_FILE}
echo "arch=('x86_64')" >> ${PKG_BUILD_FILE}
echo 'url="https://agustinsrg.github.io/pmv-site/"' >> ${PKG_BUILD_FILE}
echo "license=('MIT')" >> ${PKG_BUILD_FILE}
echo "depends=('ffmpeg')" >> ${PKG_BUILD_FILE}
echo "source=('${PMV_TAR_NAME}')" >> ${PKG_BUILD_FILE}

echo "" >> ${PKG_BUILD_FILE}

echo "package() {" >> ${PKG_BUILD_FILE}
echo '    cd "$srcdir/$pkgname-$pkgver"' >> ${PKG_BUILD_FILE}
echo '    install -Dm755 usr/bin/pmvd "${pkgdir}/usr/bin/pmvd"' >> ${PKG_BUILD_FILE}
echo '    install -Dm755 usr/bin/pmv "${pkgdir}/usr/bin/pmv"' >> ${PKG_BUILD_FILE}
echo '    install -Dm755 usr/bin/pmv-backup "${pkgdir}/usr/bin/pmv-backup"' >> ${PKG_BUILD_FILE}
echo '    install -Dm755 usr/bin/pmv-gui "${pkgdir}/usr/bin/pmv-gui"' >> ${PKG_BUILD_FILE}
echo '    install -Dm755 usr/share/pixmaps/pmv.svg "${pkgdir}/usr/share/pixmaps/pmv.svg"' >> ${PKG_BUILD_FILE}
echo '    install -Dm755 usr/share/applications/pmv.desktop "${pkgdir}/usr/share/applications/pmv.desktop"' >> ${PKG_BUILD_FILE}
echo '    mkdir -p "${pkgdir}/usr/lib/pmv"' >> ${PKG_BUILD_FILE}
echo '    cp -rf usr/lib/pmv/www "${pkgdir}/usr/lib/pmv/www"' >> ${PKG_BUILD_FILE}
echo "}" >> ${PKG_BUILD_FILE}

echo "" >> ${PKG_BUILD_FILE}

echo "Building package..."

pkgctl build

echo "DONE!"


