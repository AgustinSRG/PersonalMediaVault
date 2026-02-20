// Internationalization module

"use strict";

import type { App, Ref } from "vue";
import { nextTick, ref } from "vue";
import { addAppEventListener, emitAppEvent, EVENT_NAME_LOADED_LOCALE, EVENT_NAME_LOCALE_CHANGED } from "@/control/app-events";
import { clearLocalStorage, fetchFromLocalStorageCache, saveIntoLocalStorage } from "./utils/local-storage";

declare module "vue" {
    interface ComponentCustomProperties {
        /**
         * Current locale
         */
        $locale: Ref<string>;

        /**
         * Translates text
         * @param key The text to translate
         * @returns The translated text
         */
        $t: (key: string) => string;
    }
}

/**
 * Available languages
 * Add here when adding new languages
 */
export const AVAILABLE_LANGUAGES: { id: string; name: string }[] = [
    {
        id: "ca",
        name: "Català (Catalan)",
    },
    {
        id: "de",
        name: "Deutsch (German)",
    },
    {
        id: "el",
        name: "Ελληνικά (Greek)",
    },
    {
        id: "en",
        name: "English (English)",
    },
    {
        id: "es",
        name: "Español (Spanish)",
    },
    {
        id: "fr",
        name: "Français (French)",
    },
    {
        id: "hi",
        name: "हिंदी (Hindi)",
    },
    {
        id: "it",
        name: "Italiano (Italian)",
    },
    {
        id: "ja",
        name: "日本語 (Japanese)",
    },
    {
        id: "pt",
        name: "Português (Portuguese)",
    },
    {
        id: "ru",
        name: "Русский (Russian)",
    },
    {
        id: "zh",
        name: "中国人 (Chinese)",
    },
];

/**
 * List of supported locales
 */
export const SUPPORTED_LOCALES = AVAILABLE_LANGUAGES.map((l) => l.id);

/**
 * Base locale. This locale does not need any translation and the keys are written in this language.
 */
const BASE_LOCALE = "en";

/**
 * Fallback locale
 */
const FALLBACK_LOCALE = import.meta.env.VITE__I18N_LOCALE || BASE_LOCALE;

/**
 * Finds locale by comparing the prefix
 * @param languages List of languages
 * @param locale The locale
 * @returns The index found, or -1
 */
function findLocaleByPrefix(languages: readonly string[], locale: string): number {
    const localePrefix = locale.split("-")[0];
    for (let i = 0; i < languages.length; i++) {
        const langPrefix = languages[i].split("-")[0];

        if (langPrefix === localePrefix) {
            return i;
        }
    }

    return -1;
}

/**
 * Detects navigator language and chooses the best available locale
 * @returns The best available locale
 */
export function detectNavigatorLanguage(): string {
    const navigatorLanguages = navigator.languages || [FALLBACK_LOCALE];

    const localesSorted = SUPPORTED_LOCALES.sort((a, b) => {
        const iA = navigatorLanguages.indexOf(a);
        const iB = navigatorLanguages.indexOf(b);

        if (iA === -1 && iB === -1) {
            const jA = findLocaleByPrefix(navigatorLanguages, a);
            const jB = findLocaleByPrefix(navigatorLanguages, b);

            if (jA === -1 && jB === -1) {
                if (a === FALLBACK_LOCALE) {
                    return -1;
                } else {
                    return 1;
                }
            } else if (jA === -1) {
                return 1;
            } else if (jB === -1) {
                return -1;
            } else if (jA < jB) {
                return -1;
            } else {
                return 1;
            }
        } else if (iA === -1) {
            return 1;
        } else if (iB === -1) {
            return -1;
        } else if (iA < iB) {
            return -1;
        } else {
            return 1;
        }
    });

    return localesSorted[0] || FALLBACK_LOCALE;
}

/**
 * Key to store language preference on locale storage
 */
const LS_KEY_LANGUAGE = "app-pref-lang";

/**
 * Gets the language
 * @returns Language
 */
export function getLanguage(): string {
    return fetchFromLocalStorageCache(LS_KEY_LANGUAGE, detectNavigatorLanguage());
}

/**
 * Sets the language
 * @param lang Language
 */
export function setLanguage(lang: string) {
    saveIntoLocalStorage(LS_KEY_LANGUAGE, lang);
    emitAppEvent(EVENT_NAME_LOCALE_CHANGED, lang);
}

/**
 * Clears language setting, changing to the browser language
 */
export function clearLanguageSetting() {
    clearLocalStorage(LS_KEY_LANGUAGE);
    emitAppEvent(EVENT_NAME_LOCALE_CHANGED, detectNavigatorLanguage());
}

// Load default language

let defaultLanguage = getLanguage();

if (!SUPPORTED_LOCALES.includes(defaultLanguage)) {
    defaultLanguage = FALLBACK_LOCALE;
}

/**
 * Internationalization data
 */
export const i18nData: {
    /**
     * Locale name
     */
    locale: string;

    /**
     * Message mapping
     */
    messages: Map<string, string>;
} = {
    locale: "",
    messages: new Map(),
};

/**
 * Locale reference
 */
export const localeRef = ref("");

/**
 * Translated a message
 * @param key The translation key
 * @param locale The locale
 * @returns The translated message
 */
export function translate(key: string, locale: string): string {
    if (locale !== i18nData.locale) {
        return key;
    }
    if (i18nData.messages.has(key)) {
        return (i18nData.messages.get(key) || key).trim();
    } else {
        return key;
    }
}

/**
 * Internationalization plugin
 */
export const i18n = {
    install: (app: App) => {
        app.config.globalProperties.$locale = localeRef;

        app.config.globalProperties.$t = (key) => {
            return translate(key, localeRef.value);
        };
    },
};

/**
 * Sets page language
 * @param locale
 */
function setI18nLanguage(locale: string) {
    i18nData.locale = locale;
    localeRef.value = locale;

    document.querySelector("html").setAttribute("lang", locale);

    emitAppEvent(EVENT_NAME_LOADED_LOCALE, locale);
}

/**
 * Loads locale file
 * @param locale Locale name
 */
async function loadLocaleMessages(locale: string) {
    // Load locale messages with dynamic import
    const messages = await import(`./locales/locale-${locale}.json`);

    // Set messages
    i18nData.messages = new Map(Object.entries(messages.default));

    return nextTick();
}

/**
 * Sets and loads locale
 * @param locale The locale name
 */
async function setLocale(locale: string) {
    if (!SUPPORTED_LOCALES.includes(locale)) {
        return;
    }

    // Load locale messages
    if (locale === BASE_LOCALE) {
        // Empty map for base locale
        i18nData.messages = new Map();

        // Wait 1 animation tick
        await nextTick();
    } else {
        await loadLocaleMessages(locale);
    }

    // Set i18n language
    setI18nLanguage(locale);
}

/**
 * Loading status of the locale
 */
const LOCALE_LOAD_STATUS = {
    /**
     * True if loading
     */
    loading: false,

    /**
     * The requested locale
     */
    requested: "",
};

/**
 * Handles locale changed event
 * @param locale The locale
 */
function handleLocaleChanged(locale: string) {
    if (LOCALE_LOAD_STATUS.loading) {
        LOCALE_LOAD_STATUS.requested = locale;
        return;
    }
    LOCALE_LOAD_STATUS.loading = true;
    LOCALE_LOAD_STATUS.requested = locale;
    setLocale(locale)
        .then(() => {
            LOCALE_LOAD_STATUS.loading = false;
            if (LOCALE_LOAD_STATUS.requested !== locale) {
                handleLocaleChanged(LOCALE_LOAD_STATUS.requested);
            }
        })
        .catch((err) => {
            console.error(err);
            LOCALE_LOAD_STATUS.loading = false;
            if (LOCALE_LOAD_STATUS.requested !== locale) {
                handleLocaleChanged(LOCALE_LOAD_STATUS.requested);
            }
        });
}

addAppEventListener(EVENT_NAME_LOCALE_CHANGED, handleLocaleChanged);

emitAppEvent(EVENT_NAME_LOCALE_CHANGED, defaultLanguage);
