// Script to update the version number
// Usage: node update-version.js A.B.C

"use strict";

const FS = require("fs");
const Path = require("path");

function updateFile(file, callback) {
    const contents = FS.readFileSync(file).toString();
    const newContents = callback(contents);
    FS.writeFileSync(file, newContents);
    console.log("UPDATED: " + file);
}

function main() {
    const arg = process.argv[2] || "";

    if (!arg) {
        console.log("Usage: node update-version.js A.B.C");
        return 1;
    }

    const parts = arg.split(".");

    if (parts.length !== 3) {
        console.log("Usage: node update-version.js A.B.C");
        return 1;
    }

    const MAJOR = parseInt(parts[0], 10);
    const MINOR = parseInt(parts[1], 10);
    const REVISION = parseInt(parts[2], 10);

    if (isNaN(MAJOR) || isNaN(MINOR) || isNaN(REVISION) || MAJOR < 0 || MINOR < 0 || REVISION < 0) {
        console.log("Usage: node update-version.js A.B.C");
        return 1;
    }

    const VERSION = MAJOR + "." + MINOR + "." + REVISION;

    console.log(`Changing version to ${VERSION}`);

    updateFile(Path.resolve(__dirname, "frontend", ".env"), contents => {
        return contents
            .replace(/VITE\_\_VERSION=[0-9]+\.[0-9]+\.[0-9]+/, `VITE__VERSION=${VERSION}`);
    });

    updateFile(Path.resolve(__dirname, "frontend", "package.json"), contents => {
        return contents
            .replace(/\"version\"\: \"[0-9]+\.[0-9]+\.[0-9]+\"/, `"version": "${VERSION}"`);
    });


    updateFile(Path.resolve(__dirname, "backend", "main.go"), contents => {
        return contents
            .replace(/BACKEND_VERSION = \"[0-9]+\.[0-9]+\.[0-9]+\"/, `BACKEND_VERSION = "${VERSION}"`);
    });

    updateFile(Path.resolve(__dirname, "backend", "doc", "api-docs.yml"), contents => {
        return contents
            .replace(/version\: [0-9]+\.[0-9]+\.[0-9]+/, `version: ${VERSION}`);
    });

    updateFile(Path.resolve(__dirname, "backup-tool", "winres", "winres.json"), contents => {
        return contents
            .replace(/\"version\":\s\"[0-9]+\.[0-9]+\.[0-9]+\"/, `"version": "${VERSION}"`)
            .replace(/\"ProductVersion\":\s\"[0-9]+\.[0-9]+\.[0-9]+\"/, `"ProductVersion": "${VERSION}"`)
            .replace(/\"file_version\":\s\"[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\"/, `"file_version": "${VERSION}.0"`)
            .replace(/\"product_version\":\s\"[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\"/, `"product_version": "${VERSION}.0"`);
    });

    updateFile(Path.resolve(__dirname, "launcher", "main.go"), contents => {
        return contents
            .replace(/VERSION = \"[0-9]+\.[0-9]+\.[0-9]+\"/, `VERSION = "${VERSION}"`);
    });

    updateFile(Path.resolve(__dirname, "launcher", "winres", "winres.json"), contents => {
        return contents
            .replace(/\"version\":\s\"[0-9]+\.[0-9]+\.[0-9]+\"/, `"version": "${VERSION}"`)
            .replace(/\"ProductVersion\":\s\"[0-9]+\.[0-9]+\.[0-9]+\"/, `"ProductVersion": "${VERSION}"`)
            .replace(/\"file_version\":\s\"[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\"/, `"file_version": "${VERSION}.0"`)
            .replace(/\"product_version\":\s\"[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\"/, `"product_version": "${VERSION}.0"`);
    });

    updateFile(Path.resolve(__dirname, "launcher-gui", "Cargo.toml"), contents => {
        return contents
            .replace(/version = \"[0-9]+\.[0-9]+\.[0-9]+\"/, `version = "${VERSION}"`);
    });

    updateFile(Path.resolve(__dirname, "packages", "windows-msi", "make-wix.bat"), contents => {
        return contents
            .replace(/PersonalMediaVault\-[0-9]+\.[0-9]+\.[0-9]+\-x64\.msi/, `PersonalMediaVault-${VERSION}-x64.msi`)
            .replace(/PersonalMediaVault\-[0-9]+\.[0-9]+\.[0-9]+\-x64\-es\.msi/, `PersonalMediaVault-${VERSION}-x64-es.msi`);
    });

    updateFile(Path.resolve(__dirname, "packages", "windows-msi", "Product.wxs"), contents => {
        return contents
            .replace(/Name=\"PersonalMediaVault\" Version=\"[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+\"/, `Name="PersonalMediaVault" Version="${VERSION}.0"`);
    });

    updateFile(Path.resolve(__dirname, "packages", "dpkg-deb", "build.sh"), contents => {
        return contents
            .replace(/PMV\_VERSION\_MAJOR=[0-9]+/, `PMV_VERSION_MAJOR=${MAJOR}`)
            .replace(/PMV\_VERSION\_MINOR=[0-9]+/, `PMV_VERSION_MINOR=${MINOR}`)
            .replace(/PMV\_VERSION\_REVISION=[0-9]+/, `PMV_VERSION_REVISION=${REVISION}`);
    });

    updateFile(Path.resolve(__dirname, "packages", "rpm", "build.sh"), contents => {
        return contents
            .replace(/PMV\_VERSION\_MAJOR=[0-9]+/, `PMV_VERSION_MAJOR=${MAJOR}`)
            .replace(/PMV\_VERSION\_MINOR=[0-9]+/, `PMV_VERSION_MINOR=${MINOR}`)
            .replace(/PMV\_VERSION\_REVISION=[0-9]+/, `PMV_VERSION_REVISION=${REVISION}`);
    });

    updateFile(Path.resolve(__dirname, "packages", "pacman", "build.sh"), contents => {
        return contents
            .replace(/PMV\_VERSION\_MAJOR=[0-9]+/, `PMV_VERSION_MAJOR=${MAJOR}`)
            .replace(/PMV\_VERSION\_MINOR=[0-9]+/, `PMV_VERSION_MINOR=${MINOR}`)
            .replace(/PMV\_VERSION\_REVISION=[0-9]+/, `PMV_VERSION_REVISION=${REVISION}`);
    });

    updateFile(Path.resolve(__dirname, "packages", "linux", "build.sh"), contents => {
        return contents
            .replace(/PMV\_VERSION\_MAJOR=[0-9]+/, `PMV_VERSION_MAJOR=${MAJOR}`)
            .replace(/PMV\_VERSION\_MINOR=[0-9]+/, `PMV_VERSION_MINOR=${MINOR}`)
            .replace(/PMV\_VERSION\_REVISION=[0-9]+/, `PMV_VERSION_REVISION=${REVISION}`);
    });

    console.log("DONE!");
}

process.exit(main() || 0);
