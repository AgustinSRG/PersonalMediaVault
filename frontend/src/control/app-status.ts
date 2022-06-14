// App status
// Controls what page the app is in
// Event: app-status-update

import { getParameterByName } from "@/utils/cookie";
import { GenerateURIQuery } from "@/utils/request";
import { AppEvents } from "./app-events";

export class AppStatus {

    public static CurrentLayout = "initial";
    public static CurrentFocus = "content";

    public static CurrentPage = "home";

    public static CurrentSearch = "";

    public static ListSplitMode = true;

    public static CurrentMedia = -1;

    public static CurrentAlbum = -1;

    public static Initialize() {
        window.onpopstate = function () {
            AppStatus.GetURLParams();
            AppStatus.UpdateLayout();
            AppEvents.Emit("app-status-update", AppStatus);
        };

        AppEvents.AddEventListener("app-status-update", AppStatus.UpdateURL);

        AppStatus.GetURLParams();
        AppStatus.UpdateLayout();

        AppEvents.Emit("app-status-update", AppStatus);
    }

    public static GetURLParams() {
        const media = getParameterByName("media");
        if (media) {
            const mediaId = parseInt(media);

            if (!isNaN(mediaId) && mediaId >= 0) {
                AppStatus.CurrentMedia = mediaId;
            } else {
                AppStatus.CurrentMedia = -1;
            }
        } else {
            AppStatus.CurrentMedia = -1;
        }

        const album = getParameterByName("album");
        if (album) {
            const albumId = parseInt(album);

            if (!isNaN(albumId) && albumId >= 0) {
                AppStatus.CurrentAlbum = albumId;
            } else {
                AppStatus.CurrentAlbum = -1;
            }
        } else {
            AppStatus.CurrentAlbum = -1;
        }

        const page = getParameterByName("page");

        if (page && (["home", "random", "albums", "upload"].includes(page))) {
            AppStatus.CurrentPage = page;
        } else {
            AppStatus.CurrentPage = "home";
        }

        const search = getParameterByName("search");

        if (search) {
            AppStatus.CurrentSearch = search;
        } else {
            AppStatus.CurrentSearch = "";
        }

        const split = getParameterByName("split");

        AppStatus.ListSplitMode = (split === "yes");
    }

    public static UpdateLayout() {
        if (AppStatus.CurrentMedia >= 0) {
            if (AppStatus.CurrentAlbum >= 0) {
                // Media with album list
                AppStatus.CurrentLayout = "album";
            } else if (AppStatus.ListSplitMode) {
                // Media with list
                AppStatus.CurrentLayout = "media-split";
            } else {
                // Media alone
                AppStatus.CurrentLayout = "media";
            }
        } else if (AppStatus.CurrentAlbum >= 0) {
            AppStatus.CurrentLayout = "album";
        } else {
            // Just initial, sidebar + List
            AppStatus.CurrentLayout = "initial";
        }
    }

    public static GetCurrentURL(): string {
        const params: any = Object.create(null);

        if (AppStatus.CurrentMedia >= 0) {
            params["media"] = AppStatus.CurrentMedia + "";
        }

        if (AppStatus.CurrentAlbum >= 0) {
            params["album"] = AppStatus.CurrentAlbum + "";
        }

        if (AppStatus.CurrentPage && AppStatus.CurrentPage !== "home") {
            params["page"] = AppStatus.CurrentPage;
        }

        if (AppStatus.CurrentSearch) {
            params["search"] = AppStatus.CurrentSearch;
        }

        if (AppStatus.ListSplitMode) {
            params["split"] = "yes";
        }

        return window.location.protocol + "//" + window.location.host + window.location.pathname + GenerateURIQuery(params);
    }

    public static UpdateURL() {
        if (history.pushState) {
            const newurl = AppStatus.GetCurrentURL();
            // Update URL
            if (newurl !== location.href) {
                window.history.pushState({ path: newurl }, '', newurl);
            }
        }
    }

    public static OnStatusUpdate() {
        AppStatus.UpdateLayout();

        AppEvents.Emit("app-status-update", AppStatus);

        AppStatus.UpdateURL();
    }

    public static GoToPage(page: string) {
        AppStatus.CurrentPage = page;

        AppStatus.CurrentAlbum = -1;
        AppStatus.CurrentMedia = -1;

        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "content";

        AppStatus.OnStatusUpdate();
    }

    public static GoToSearch(search: string) {
        AppStatus.CurrentPage = "home";
        AppStatus.CurrentSearch = search;

        AppStatus.CurrentAlbum = -1;
        AppStatus.CurrentMedia = -1;

        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "content";

        AppStatus.OnStatusUpdate();
    }

    public static ClickOnMedia(mediaId: number) {
        AppStatus.CurrentAlbum = -1;
        AppStatus.CurrentMedia = mediaId;

        AppStatus.ListSplitMode = true;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "content";

        AppStatus.OnStatusUpdate();
    }

    public static ClickOnAlbum(albumId: number, mediaId: number) {
        AppStatus.CurrentAlbum = albumId;
        AppStatus.CurrentMedia = mediaId;

        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "content";

        AppStatus.OnStatusUpdate();
    }

    public static FocusAltContent() {
        AppStatus.CurrentFocus = "alt";

        AppStatus.OnStatusUpdate();
    }

    public static OpenMenu() {
        AppStatus.CurrentFocus = "menu";

        AppStatus.OnStatusUpdate();
    }

    public static CloseMenu() {
        AppStatus.CurrentFocus = "content";

        AppStatus.OnStatusUpdate();
    }
}
