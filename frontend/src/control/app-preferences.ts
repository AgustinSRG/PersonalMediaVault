// App preferences

import { AlbumsController } from "./albums";
import { AppEvents } from "./app-events";
import { fetchFromLocalStorage, saveIntoLocalStorage } from "../utils/local-storage";
import { AlbumListItemMin } from "@/api/models";

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

    public static PageMaxItems = 25;

    public static PageItemsSize = "normal";
    public static PageItemsFit = 0;

    public static LoadPreferences() {
        const locale = navigator.language || "en";

        AppPreferences.Language = fetchFromLocalStorage("app-pref-lang", locale);
        AppPreferences.Theme = fetchFromLocalStorage("app-pref-theme", defaultBrowserTheme());
        AppPreferences.AlbumPositionMap = fetchFromLocalStorage("app-pref-albums-order", Object.create(null));
        AppPreferences.FavAlbums = fetchFromLocalStorage("app-pref-albums-fav", []);

        AppPreferences.PageMaxItems = fetchFromLocalStorage("app-pref-page-max-items", 25);

        AppPreferences.PageItemsSize = fetchFromLocalStorage("app-pref-page-items-size", "normal");
        AppPreferences.PageItemsFit = fetchFromLocalStorage("app-pref-page-items-fit", 5);

        AppEvents.AddEventListener("albums-update", AppPreferences.OnAlbumsUpdate);
        AppEvents.AddEventListener("current-album-update", AppPreferences.OnAlbumLoad);
    }

    public static SetLanguage(lang: string) {
        AppPreferences.Language = lang;
        saveIntoLocalStorage("app-pref-lang", lang);
    }

    public static SetTheme(t: string) {
        AppPreferences.Theme = t;
        saveIntoLocalStorage("app-pref-theme", t);
        AppEvents.Emit("theme-changed", t);
    }

    public static OnAlbumsUpdate(albums: { [id: string]: AlbumListItemMin }) {
        for (const id of Object.keys(AppPreferences.AlbumPositionMap)) {
            if (!albums[id]) {
                delete AppPreferences.AlbumPositionMap[id];
            }
        }
        AppPreferences.FavAlbums = AppPreferences.FavAlbums.filter((id) => {
            return !!albums[id];
        });
        saveIntoLocalStorage("app-pref-albums-order", AppPreferences.AlbumPositionMap);
        saveIntoLocalStorage("app-pref-albums-fav", AppPreferences.FavAlbums);
    }

    public static OnAlbumLoad() {
        if (!AlbumsController.CurrentAlbumData) {
            return;
        }
        AppPreferences.AlbumPositionMap[AlbumsController.CurrentAlbumData.id + ""] = Date.now();
        saveIntoLocalStorage("app-pref-albums-order", AppPreferences.AlbumPositionMap);
        AppEvents.Emit("album-sidebar-top", AlbumsController.CurrentAlbumData.id);
    }

    public static albumAddFav(id: string) {
        if (!AppPreferences.FavAlbums.includes(id)) {
            AppPreferences.FavAlbums.push(id);
            saveIntoLocalStorage("app-pref-albums-fav", AppPreferences.FavAlbums);
            AppEvents.Emit("albums-fav-updated");
        }
    }

    public static albumRemoveFav(id: string) {
        const index = AppPreferences.FavAlbums.indexOf(id);
        if (index >= 0) {
            AppPreferences.FavAlbums.splice(index, 1);
            saveIntoLocalStorage("app-pref-albums-fav", AppPreferences.FavAlbums);
            AppEvents.Emit("albums-fav-updated");
        }
    }

    public static SetPageMaxItems(m: number) {
        AppPreferences.PageMaxItems = m;
        saveIntoLocalStorage("app-pref-page-max-items", m);
        AppEvents.Emit("page-size-pref-updated");
    }

    public static SetPageItemsSize(s: string) {
        AppPreferences.PageItemsSize = s;
        saveIntoLocalStorage("app-pref-page-items-size", s);
        AppEvents.Emit("page-items-pref-updated");
    }

    public static SetPageItemsFit(m: number) {
        AppPreferences.PageItemsFit = m;
        saveIntoLocalStorage("app-pref-page-items-fit", m);
        AppEvents.Emit("page-items-pref-updated");
    }
}
