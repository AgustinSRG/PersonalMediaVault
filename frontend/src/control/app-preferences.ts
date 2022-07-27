// App preferences

import { AlbumEntry, AlbumsController } from "./albums";
import { AppEvents } from "./app-events";
import { LocalStorage } from "./local-storage";

export class AppPreferences {
    public static Language = "en";
    public static Theme = "dark";
    public static AlbumPositionMap: {[id: string]: number} = Object.create(null);


    public static LoadPreferences() {
        const locale = (navigator.language || "en");

        AppPreferences.Language = LocalStorage.Get("app-pref-lang", locale);
        AppPreferences.Theme = LocalStorage.Get("app-pref-theme", "light");
        AppPreferences.AlbumPositionMap = LocalStorage.Get("app-pref-albums-order", Object.create(null));

        AppEvents.AddEventListener("albums-update", AppPreferences.OnAlbumsUpdate);
        AppEvents.AddEventListener("current-album-update", AppPreferences.OnAlbumLoad);
    }

    public static SetLanguage(lang: string) {
        AppPreferences.Language = lang;
        LocalStorage.Set("app-pref-lang", lang);
    }

    public static SetTheme(t: string) {
        AppPreferences.Theme = t;
        LocalStorage.Set("app-pref-theme", t);
        AppEvents.Emit("theme-changed", t);
    }

    public static OnAlbumsUpdate(albums: { [id: string]: AlbumEntry }) {
        for (const id of Object.keys(AppPreferences.AlbumPositionMap)) {
            if (!albums[id]) {
                delete AppPreferences.AlbumPositionMap[id];
            }
        }
        LocalStorage.Set("app-pref-albums-order", AppPreferences.AlbumPositionMap);
    }

    public static OnAlbumLoad() {
        if (!AlbumsController.CurrentAlbumData) {
            return;
        }
        AppPreferences.AlbumPositionMap[AlbumsController.CurrentAlbumData.id + ""] = Date.now();
        LocalStorage.Set("app-pref-albums-order", AppPreferences.AlbumPositionMap);
        AppEvents.Emit("album-sidebar-top", AlbumsController.CurrentAlbumData.id);
    }
}
