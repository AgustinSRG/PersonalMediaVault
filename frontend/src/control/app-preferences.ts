// App preferences

import { LocalStorage } from "./local-storage";

export class AppPreferences {
    public static Language = "en";
    public static Theme = "dark";


    public static LoadPreferences() {
        AppPreferences.Language = LocalStorage.Get("app-pref-lang", "");
        AppPreferences.Theme = LocalStorage.Get("app-pref-theme", "");
    }

    public static SetLanguage(lang: string) {
        AppPreferences.Language = lang;
        LocalStorage.Set("app-pref-lang", lang);
    }

    public static SetTheme(t: string) {
        AppPreferences.Theme = t;
        LocalStorage.Set("app-pref-theme", t);
    }
}
