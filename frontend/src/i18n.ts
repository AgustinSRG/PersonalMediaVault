// Internationalization module

"use strict";

import { App, Ref, nextTick, ref } from "vue";
import { AppEvents } from "./control/app-events";
import { EVENT_NAME_LOCALE_CHANGED, getLanguage } from "./control/app-preferences";
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

        /**
         * Updates locale
         * @param locale The new locale
         */
        $updateLocale: (locale: string) => void;
    }
}

/**
 * List of supported locales
 */
export const SUPPORTED_LOCALES = ["en", "es"];

// Load default language

let defaultLanguage = getLanguage();
const fallbackLocale = import.meta.env.VITE__I18N_FALLBACK_LOCALE || "en";

if (!SUPPORTED_LOCALES.includes(defaultLanguage)) {
    defaultLanguage = fallbackLocale;
}

/**
 * Internationalization data
 */
const i18nData: {
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
 * Internationalization plugin
 */
export const i18n = {
    install: (app: App) => {
        app.config.globalProperties.$locale = ref("");
        app.config.globalProperties.$updateLocale = (locale: string) => {
            app.config.globalProperties.$locale.value = locale;
        };

        app.config.globalProperties.$t = (key) => {
            if (app.config.globalProperties.$locale.value !== i18nData.locale) {
                return key;
            }
            if (i18nData.messages.has(key)) {
                return (i18nData.messages.get(key) || key).trim();
            } else {
                return key;
            }
        };
    },
};

/**
 * Event triggered when a new locale file is loaded
 */
export const EVENT_NAME_LOADED_LOCALE = "loaded-locale";

/**
 * Sets page language
 * @param locale
 */
function setI18nLanguage(locale: string) {
    i18nData.locale = locale;

    document.querySelector("html").setAttribute("lang", locale);

    AppEvents.Emit(EVENT_NAME_LOADED_LOCALE, locale);
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
    await loadLocaleMessages(locale);

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

AppEvents.AddEventListener(EVENT_NAME_LOCALE_CHANGED, handleLocaleChanged);

AppEvents.Emit(EVENT_NAME_LOCALE_CHANGED, defaultLanguage);
