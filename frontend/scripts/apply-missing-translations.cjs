// Script to apply the missing translations that are ready

"use strict";

const FS = require("fs");
const Path = require("path");

function findTranslationConfigFiles() {
    const files = FS.readdirSync(Path.resolve(__dirname, "..", "src", "locales"));
    const result = [];
    for (let file of files) {
        if (file.endsWith(".json")) {
            result.push(Path.resolve(__dirname, "..", "src", "locales", file));
        }
    }
    return result;
}

function isReady(lines) {
    for (let line of lines) {
        if (line === "STATUS=READY") {
            return true;
        }

        if (line === "STATUS=PENDING") {
            return false;
        }

        if (line === "--- BEGIN TRANSLATION ---") {
            return false;
        }
    }

    return false;
}

function getMissingTranslations(lines) {
    const res = [];
    let started = false;

    for (let line of lines) {
        if (started) {
            if (line === "--- END TRANSLATION ---") {
                break;
            } else if (line) {
                res.push(line);
            }
        } else if (line === "--- BEGIN TRANSLATION ---") {
            started = true;
        }
    }

    return res;
}

function main() {
    console.log("Applying missing translations...");

    const locales = findTranslationConfigFiles().map((l) => {
        const lang = Path.basename(l).split("-").pop().split(".")[0];
        return {
            id: lang,
            file: l,
            missingFile: Path.resolve(__dirname, "..", "src", "locales", "locale-" + lang + ".missing.txt"),
        };
    });

    for (let locale of locales) {
        if (locale.id === "en") {
            continue;
        }

        if (!FS.existsSync(locale.missingFile)) {
            continue;
        }

        const missingLines = FS.readFileSync(locale.missingFile)
            .toString()
            .split("\n")
            .map((l) => {
                return l.replace(/\r/g, "");
            });

        if (!isReady(missingLines)) {
            console.log(`Missing translations for locale ${locale.id} are still pending.`);
            continue;
        }

        console.log(`Missing translations for locale ${locale.id} are ready. Applying...`);

        const preparedMissingTranslations = getMissingTranslations(missingLines);

        const locales = JSON.parse(FS.readFileSync(locale.file).toString());
        let missingCount = 0;

        for (let key of Object.keys(locales)) {
            if (!locales[key]) {
                // Missing translation

                if (preparedMissingTranslations.length === 0) {
                    console.log(`Warning: The locale file has more missing translations than provided.`);
                    break;
                }

                locales[key] = preparedMissingTranslations.shift();
                missingCount++;
            }
        }

        if (preparedMissingTranslations.length > 0) {
            console.log(`Warning: The locale file has less missing translations than provided.`);
        }

        FS.writeFileSync(locale.file, JSON.stringify(locales, null, 4) + "\n");
        console.log(`Updated: ${locale.file}`);

        FS.unlinkSync(locale.missingFile);
        console.log(`Removed: ${locale.missingFile}`);

        console.log(`Done: Filled ${missingCount} missing translations for locale ${locale.id}.`);
    }
}

main();
