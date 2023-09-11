import { App, Ref, nextTick, ref } from "vue";
import { AppEvents } from "./control/app-events";
import { LocalStorage } from "./control/local-storage";

declare module "vue" {
    interface ComponentCustomProperties {
        $locale: Ref<string>;
        $t: (key: string) => string;
        $updateLocale: (locale: string) => void;
    }
}

export const SUPPORT_LOCALES = ["en", "es"];

let defaultLanguage = LocalStorage.Get("app-pref-lang", import.meta.env.VITE__I18N_LOCALE || "en");
const fallbackLocale = import.meta.env.VITE__I18N_FALLBACK_LOCALE || "en";

if (!SUPPORT_LOCALES.includes(defaultLanguage)) {
    defaultLanguage = fallbackLocale;
}

const i18nData: {
    locale: string;
    messages: Map<string, string>;
} = {
    locale: "",
    messages: new Map(),
};

// Plugin
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
                return i18nData.messages.get(key);
            } else {
                return key;
            }
        };
    },
};

function setI18nLanguage(locale: string) {
    i18nData.locale = locale;

    document.querySelector("html").setAttribute("lang", locale);

    AppEvents.Emit("loaded-locale", locale);
}

async function loadLocaleMessages(locale: string) {
    // load locale messages with dynamic import
    const messages = await import(/* webpackChunkName: "locale-[request]" */ `./locales/${locale}.json`);

    // Set messages
    i18nData.messages = new Map(Object.entries(messages.default));

    return nextTick();
}

async function setLocale(locale: string) {
    // use locale if paramsLocale is not in SUPPORT_LOCALES
    if (!SUPPORT_LOCALES.includes(locale)) {
        return;
    }

    // load locale messages
    await loadLocaleMessages(locale);

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
