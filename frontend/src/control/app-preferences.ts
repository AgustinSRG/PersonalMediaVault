// App preferences

import { AlbumEntry, AlbumsController } from "./albums";
import { AppEvents } from "./app-events";
import { LocalStorage } from "./local-storage";

function defaultBrowserTheme(): string {
    if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
        return "dark";
    } else if (window.matchMedia) {
        return "light";
    } else {
        return "dark";
    }
}

export class AppPreferences {
    public static Language = "en";
    public static Theme = "dark";
    public static AlbumPositionMap: { [id: string]: number } = Object.create(null);
    public static FavAlbums: string[] = [];

    public static LoadPreferences() {
        const locale = navigator.language || "en";

        AppPreferences.Language = LocalStorage.Get("app-pref-lang", locale);
        AppPreferences.Theme = LocalStorage.Get("app-pref-theme", defaultBrowserTheme());
        AppPreferences.AlbumPositionMap = LocalStorage.Get("app-pref-albums-order", Object.create(null));
        AppPreferences.FavAlbums = LocalStorage.Get("app-pref-albums-fav", []);

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
        AppPreferences.FavAlbums = AppPreferences.FavAlbums.filter((id) => {
            return !!albums[id];
        });
        LocalStorage.Set("app-pref-albums-order", AppPreferences.AlbumPositionMap);
        LocalStorage.Set("app-pref-albums-fav", AppPreferences.FavAlbums);
    }

    public static OnAlbumLoad() {
        if (!AlbumsController.CurrentAlbumData) {
            return;
        }
        AppPreferences.AlbumPositionMap[AlbumsController.CurrentAlbumData.id + ""] = Date.now();
        LocalStorage.Set("app-pref-albums-order", AppPreferences.AlbumPositionMap);
        AppEvents.Emit("album-sidebar-top", AlbumsController.CurrentAlbumData.id);
    }

    public static albumAddFav(id: string) {
        if (!AppPreferences.FavAlbums.includes(id)) {
            AppPreferences.FavAlbums.push(id);
            LocalStorage.Set("app-pref-albums-fav", AppPreferences.FavAlbums);
            AppEvents.Emit("albums-fav-updated");
        }
    }

    public static albumRemoveFav(id: string) {
        const index = AppPreferences.FavAlbums.indexOf(id);
        if (index >= 0) {
            AppPreferences.FavAlbums.splice(index, 1);
            LocalStorage.Set("app-pref-albums-fav", AppPreferences.FavAlbums);
            AppEvents.Emit("albums-fav-updated");
        }
    }
}
