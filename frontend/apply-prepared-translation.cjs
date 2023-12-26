// Script apply a prepared translation

"use strict";

const Path = require("path");
const FS = require("fs");

function main() {
    const lang = process.argv[2];

    if (!lang) {
        console.log("Usage: node apply-prepared-translation <locale>");
        process.exit(1);
    }

    if (lang === "en") {
        console.log(`Cannot translate to english as it is the default locale`);
        process.exit(1);
    }

    const preparedFile = Path.resolve(__dirname, "prepared-translation.txt");

    if (!FS.existsSync(preparedFile)) {
        console.log(`File not found: ${preparedFile}`);
        process.exit(1);
    }

    const localeFile = Path.resolve(__dirname, "src", "locales", "locale-" + lang + ".json");

    if (!FS.existsSync(localeFile)) {
        console.log(`File not found: ${localeFile}`);
        process.exit(1);
    }

    const preparedMissingTranslations = FS.readFileSync(preparedFile)
        .toString()
        .split("\n")
        .map((p) => {
            return p.trim();
        })
        .filter((p) => !!p);

    const locales = JSON.parse(FS.readFileSync(localeFile).toString());
    let missingCount = 0;

    for (let key of Object.keys(locales)) {
        if (!locales[key]) {
            // Missing translation

            if (preparedMissingTranslations.length === 0) {
                console.log(
                    `Error: The locale file has more missing translations than provided in prepared-translation.txt. Did you update the locales before applying?`,
                );
                process.exit(1);
            }

            locales[key] = preparedMissingTranslations.shift();
            missingCount++;
        }
    }

    if (preparedMissingTranslations.length > 0) {
        console.log(
            `Error: The locale file has less missing translations than provided in prepared-translation.txt. Did you update the locales before applying?`,
        );
        process.exit(1);
    }

    FS.writeFileSync(localeFile, JSON.stringify(locales, null, 4) + "\n");
    console.log(`Updated: ${localeFile}`);

    FS.unlinkSync(preparedFile);
    console.log(`Removed: ${preparedFile}`);

    console.log(`Done: Filled ${missingCount} missing translations for locale ${lang}.`);
}

main();
