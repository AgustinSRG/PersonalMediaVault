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

cd ../packages/rpm

# Create compressed package

PMV_VERSION_MAJOR=4
PMV_VERSION_MINOR=0
PMV_VERSION_REVISION=1

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

# Compress into a source file

echo "Building compressed source file..."

tar czf ${PMV_TAR_NAME} ${PMV_PKG_FOLDER}

mv -f ${PMV_TAR_NAME} ~/rpmbuild/SOURCES/${PMV_TAR_NAME}

echo "Built compressed source file: ${PMV_TAR_NAME}"

# Built spec file

SPEC_FILE=./${PMV_PKG_NAME}.spec

echo "Creating spec file: ${SPEC_FILE}"

echo "Name:      personalmediavault" > ${SPEC_FILE}
echo "Version:   ${PMV_VERSION_MAJOR}.${PMV_VERSION_MINOR}.${PMV_VERSION_REVISION}" >> ${SPEC_FILE}
echo "Release:   1%{?dist}" >> ${SPEC_FILE}
echo "Summary:   Personal Media Vault" >> ${SPEC_FILE}
echo "BuildArch: x86_64" >> ${SPEC_FILE}
echo "URL:       https://agustinsrg.github.io/pmv-site/" >> ${SPEC_FILE}
echo "BugURL:    https://github.com/AgustinSRG/PersonalMediaVault/issues" >> ${SPEC_FILE}
echo "License:   MIT" >> ${SPEC_FILE}
echo "Requires:  ffmpeg-free" >> ${SPEC_FILE}
echo "Source0:   ${PMV_TAR_NAME}" >> ${SPEC_FILE}

echo "" >> ${SPEC_FILE}

echo "%description" >> ${SPEC_FILE}
echo "Self-hosted web application to store media files (video, audio and pictures) in an encrypted storage, and visualize them using a web browser." >> ${SPEC_FILE}

echo "" >> ${SPEC_FILE}

echo "%global debug_package %{nil}" >> ${SPEC_FILE}

echo "" >> ${SPEC_FILE}

echo "%prep" >> ${SPEC_FILE}
echo "%setup -q" >> ${SPEC_FILE}

echo "" >> ${SPEC_FILE}

echo "%install" >> ${SPEC_FILE}
echo 'rm -rf $RPM_BUILD_ROOT' >> ${SPEC_FILE}
echo 'mkdir -p $RPM_BUILD_ROOT/usr/bin' >> ${SPEC_FILE}
echo 'mkdir -p $RPM_BUILD_ROOT/usr/lib/pmv' >> ${SPEC_FILE}
echo 'mkdir -p $RPM_BUILD_ROOT/usr/share/pixmaps/' >> ${SPEC_FILE}
echo 'mkdir -p $RPM_BUILD_ROOT/usr/share/applications' >> ${SPEC_FILE}
echo 'cp usr/bin/pmvd $RPM_BUILD_ROOT/usr/bin/pmvd' >> ${SPEC_FILE}
echo 'cp usr/bin/pmv $RPM_BUILD_ROOT/usr/bin/pmv' >> ${SPEC_FILE}
echo 'cp usr/bin/pmv-backup $RPM_BUILD_ROOT/usr/bin/pmv-backup' >> ${SPEC_FILE}
echo 'cp usr/bin/pmv-gui $RPM_BUILD_ROOT/usr/bin/pmv-gui' >> ${SPEC_FILE}
echo 'cp -r usr/lib/pmv/www $RPM_BUILD_ROOT/usr/lib/pmv/www' >> ${SPEC_FILE}
echo 'cp usr/share/pixmaps/pmv.svg $RPM_BUILD_ROOT/usr/share/pixmaps/pmv.svg' >> ${SPEC_FILE}
echo 'cp usr/share/applications/pmv.desktop $RPM_BUILD_ROOT/usr/share/applications/pmv.desktop' >> ${SPEC_FILE}

echo "" >> ${SPEC_FILE}

echo "%clean" >> ${SPEC_FILE}
echo 'rm -rf $RPM_BUILD_ROOT' >> ${SPEC_FILE}

echo "" >> ${SPEC_FILE}

echo "%files" >> ${SPEC_FILE}
echo '/usr/bin/pmvd' >> ${SPEC_FILE}
echo '/usr/bin/pmv' >> ${SPEC_FILE}
echo '/usr/bin/pmv-backup' >> ${SPEC_FILE}
echo '/usr/bin/pmv-gui' >> ${SPEC_FILE}
echo '/usr/lib/pmv/www/' >> ${SPEC_FILE}
echo '/usr/share/pixmaps/pmv.svg' >> ${SPEC_FILE}
echo '/usr/share/applications/pmv.desktop' >> ${SPEC_FILE}

mv -f ${SPEC_FILE} ~/rpmbuild/SPECS/${SPEC_FILE}

# Build RMP package

echo "Building RPM..."

rpmbuild -bb ~/rpmbuild/SPECS/${SPEC_FILE}

# Move RPM here

mv -f ~/rpmbuild/RPMS/x86_64/personalmediavault-*.rpm .

echo "DONE!"


