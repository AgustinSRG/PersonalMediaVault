// Script to detect font-awesome usages and
// making a subset using the official release

"use strict";

const Font = require("fonteditor-core").Font;
const FS = require("fs");
const Path = require("path");
const ttf2woff2 = require("ttf2woff2");

const FA_VERSION = "6.0.0";

function scanDirectories(root) {
    const files = [];
    const filesInDir = FS.readdirSync(root);

    for (const file of filesInDir) {
        const absFile = Path.resolve(root, file);
        const relFile = absFile.substr(Path.resolve(__dirname).length).substr(1);

        if (relFile === ".git") {
            continue;
        }

        const stats = FS.lstatSync(absFile);
        if (stats.isDirectory()) {
            const subFiles = scanDirectories(absFile);
            for (const sf of subFiles) {
                files.push(sf);
            }
        } else if (stats.isFile()) {
            if (file.endsWith(".vue")) {
                files.push(absFile);
            }
        }
    }

    return files;
}

function getAffectedClassesFromCssRule(css) {
    let status = 0;
    const result = [];
    let classBuf = "";
    for (let i = 0; i < css.length; i++) {
        const c = css.charAt(i);

        switch (status) {
            case 0:
                switch (c) {
                    case "@":
                        status = 1;
                        classBuf = "";
                        break;
                    case ".":
                        break;
                    case " ":
                    case "\n":
                    case "\t":
                    case "\r":
                    case ",":
                        if (classBuf) {
                            result.push(classBuf);
                        }
                        classBuf = "";
                        break;
                    case "{":
                        if (classBuf) {
                            result.push(classBuf);
                        }
                        classBuf = "";
                        status = 9;
                        break;
                    case ":":
                        status = 2;
                        break;
                    default:
                        classBuf += c;
                }
                break;
            case 1:
            case 2:
                switch (c) {
                    case " ":
                    case "\n":
                    case "\t":
                    case "\r":
                    case ",":
                        if (classBuf) {
                            result.push(classBuf);
                        }
                        classBuf = "";
                        status = 0;
                        break;
                    case "{":
                        if (classBuf) {
                            result.push(classBuf);
                        }
                        classBuf = "";
                        status = 9;
                        break;
                }
                break;
        }
    }

    if (classBuf) {
        result.push(classBuf);
    }

    return result;
}

function main() {
    console.log("Detecting font-awesome usages...");

    const filesToCheck = scanDirectories(Path.resolve(__dirname, "..", "src"));

    const usages = new Set();

    for (const file of filesToCheck) {
        const fileContents = FS.readFileSync(file).toString();

        const matches = fileContents.match(/(\s|\n|\t|\"|\')?((fas)|(far)|(fab)|(fa\-[a-z0-9\-]+))(\s|\n|\t|\"|\')/g) || [];

        for (let match of matches) {
            usages.add(match.replace(/[^a-z0-9\-]+/g, ""));
        }
    }

    console.log("Found: " + Array.from(usages).join(", "));

    console.log("Preparing css file...");

    const allFontAwesomeLines = FS.readFileSync(Path.resolve(__dirname, "..", "font-awesome", FA_VERSION, "css", "all.css"))
        .toString()
        .split("\n")
        .slice(5);

    const entries = [];
    let buf = [];

    for (const line of allFontAwesomeLines) {
        if (line.trim() === "") {
            if (buf.length > 0) {
                entries.push({
                    css: buf.join("\n"),
                    affected: getAffectedClassesFromCssRule(buf.join("\n")),
                    isMedia: buf[0].startsWith("@media"),
                    isFontFace: buf[0].startsWith("@font-face"),
                    isKeyFrames: buf[0].startsWith("@-webkit-keyframes") || buf[0].startsWith("@keyframes"),
                });
            }
            buf = [];
        } else {
            buf.push(line);
        }
    }

    let finalCssEntries = [];
    const setUnicodeChars = new Set();
    const mapUnicodeChars = new Map();

    for (let entry of entries) {
        if (!entry.isMedia && !entry.isFontFace) {
            let mustInclude = false;
            for (let a of entry.affected) {
                if (a === "fa" || usages.has(a)) {
                    mustInclude = true;
                    break;
                }
            }
            if (!mustInclude) {
                continue;
            }
        }

        if (entry.isFontFace) {
            entry.css = entry.css.replace(/\.\.\/webfonts\//g, "./");
        } else if (!entry.isMedia && !entry.isKeyFrames) {
            const cssLines = entry.css.split("\n");

            if (cssLines[0].trim().endsWith("::before {")) {
                const hexCode = ((cssLines[1] + "").split(":").pop().split('"')[1] + "").substring(1);
                const n = parseInt(hexCode, 16);
                if (!isNaN(n)) {
                    setUnicodeChars.add(n);
                    mapUnicodeChars.set(entry.affected[0] + "", n);
                }
            }
        }

        finalCssEntries.push(entry.css);
    }

    FS.writeFileSync(Path.resolve(__dirname, "..", "src", "assets", "font-awesome.css"), finalCssEntries.join("\n\n"));

    console.log("Map of affected classes: ");

    for (const c of usages) {
        const n = mapUnicodeChars.get(c);
        console.log("    " + c + " -> " + n);
    }

    console.log("Preparing font files... (TTF)");

    const fontFiles = FS.readdirSync(Path.resolve(__dirname, "..", "font-awesome", FA_VERSION, "webfonts"));

    for (const fontFile of fontFiles) {
        const format = fontFile.split(".").pop();
        if (format !== "ttf") {
            continue;
        }
        const fontFileContents = FS.readFileSync(Path.resolve(__dirname, "..", "font-awesome", FA_VERSION, "webfonts", fontFile));

        const font = Font.create(fontFileContents, {
            type: format,
            subset: Array.from(setUnicodeChars),
            hinting: true, // save font hinting
            compound2simple: true, // transform ttf compound glyf to simple
            inflate: null, // inflate function for woff
            combinePath: false, // for svg path
        });

        const finalFontBuffer = font.write({
            type: format,
        });

        FS.writeFileSync(Path.resolve(__dirname, "..", "src", "assets", fontFile), finalFontBuffer);
        console.log("WRITE: " + fontFile);
    }

    console.log("Encoding to Woff2...");

    const resultFontFiles = FS.readdirSync(Path.resolve(__dirname, "..", "src", "assets"));

    for (let fontFile of resultFontFiles) {
        if (!fontFile.endsWith(".ttf")) {
            continue;
        }

        const woff2File = fontFile.replace(".ttf", ".woff2");

        const inputData = FS.readFileSync(Path.resolve(__dirname, "..", "src", "assets", fontFile));

        FS.writeFileSync(Path.resolve(__dirname, "..", "src", "assets", woff2File), ttf2woff2(inputData));
    }

    console.log("DONE!");
}

main();
