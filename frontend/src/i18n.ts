import { nextTick } from "vue";
import { createI18n } from "vue-i18n";
import { AppEvents } from "./control/app-events";
import { LocalStorage } from "./control/local-storage";

export const SUPPORT_LOCALES = ["en", "es"];

let defaultLanguage = LocalStorage.Get("app-pref-lang", process.env.VUE_APP_I18N_LOCALE || "en");
const fallbackLocale = process.env.VUE_APP_I18N_FALLBACK_LOCALE || "en";

if (!SUPPORT_LOCALES.includes(defaultLanguage)) {
    defaultLanguage = fallbackLocale;
}

export const i18n = createI18n({
    locale: <any>"",
});

function setI18nLanguage(locale: string) {
    if (i18n.mode === "legacy") {
        i18n.global.locale = locale;
    } else {
        i18n.global.locale.value = locale;
    }
    /**
     * NOTE:
     * If you need to specify the language setting for headers, such as the `fetch` API, set it here.
     * The following is an example for axios.
     *
     * axios.defaults.headers.common['Accept-Language'] = locale
     */
    document.querySelector("html").setAttribute("lang", locale);
}

async function loadLocaleMessages(locale: string) {
    // load locale messages with dynamic import
    const messages = await import(/* webpackChunkName: "locale-[request]" */ `./locales/${locale}.json`);

    // set locale and locale message
    i18n.global.setLocaleMessage(locale, messages.default);

    return nextTick();
}

async function setLocale(locale: string) {
    // use locale if paramsLocale is not in SUPPORT_LOCALES
    if (!SUPPORT_LOCALES.includes(locale)) {
        return;
    }

    // load locale messages
    if (!i18n.global.availableLocales.includes(locale)) {
        await loadLocaleMessages(locale);
    }

    // set i18n language
    setI18nLanguage(locale);
}

const LOCALE_LOAD_STATUS = {
    loading: false,
    requested: "",
};

function handleNextEvent(locale: string) {
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
                handleNextEvent(LOCALE_LOAD_STATUS.requested);
            }
        })
        .catch((err) => {
            console.error(err);
            LOCALE_LOAD_STATUS.loading = false;
            if (LOCALE_LOAD_STATUS.requested !== locale) {
                handleNextEvent(LOCALE_LOAD_STATUS.requested);
            }
        });
}

AppEvents.AddEventListener("set-locale", handleNextEvent);

AppEvents.Emit("set-locale", defaultLanguage);
