// Script to prepare translation in a non-JSON format

"use strict";

const Path = require("path");
const FS = require("fs");

function main() {
    const lang = process.argv[2];

    if (!lang) {
        console.log("Usage: node prepare-translation <locale>");
        process.exit(1);
    }

    if (lang === "en") {
        console.log(`Cannot translate to english as it is the default locale`);
        process.exit(1);
    }

    const localeFile = Path.resolve(__dirname, "src", "locales", "locale-" + lang + ".json");

    if (!FS.existsSync(localeFile)) {
        console.log(`File not found: ${localeFile}`);
        process.exit(1);
    }

    const locales = JSON.parse(FS.readFileSync(localeFile).toString());

    const missingTranslations = [];

    for (let key of Object.keys(locales)) {
        if (!locales[key]) {
            // Missing translation
            missingTranslations.push(key);
        }
    }

    if (missingTranslations.length === 0) {
        console.log(`Locale ${lang} has no missing translations!`);
        return;
    }

    console.log(`Found ${missingTranslations.length} missing translations for locale ${lang}.`);

    FS.writeFileSync(Path.resolve(__dirname, "prepared-translation.txt"), missingTranslations.join("\n"));

    console.log("Write file: prepared-translation.txt");
    console.log(`Translate all the line of the file to the locale '${lang}' and then run:`);
    console.log(`    node apply-prepared-translation.cjs ${lang}`);
    console.log("");
    console.log("IMPORTANT: DO NOT UPDATE THE LOCALES BEFORE APPLYING OR THE ORDER WILL CHANGE");
}

main();
