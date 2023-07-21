// App status
// Controls what page the app is in
// Event: app-status-update

import { getParameterByName } from "@/utils/cookie";
import { GenerateURIQuery } from "@/utils/request";
import { AlbumsController } from "./albums";
import { AppEvents } from "./app-events";
import { PlayerPreferences } from "./player-preferences";

export class AppStatus {

    public static CurrentLayout = "initial";
    public static CurrentFocus = "left";

    public static CurrentPage = "home";

    public static CurrentSearch = "";

    public static SearchParams = "";

    public static ListSplitMode = true;

    public static CurrentMedia = -1;

    public static CurrentAlbum = -1;

    public static Initialize() {
        window.onpopstate = function () {
            AppStatus.GetURLParams();
            AppStatus.OnStatusUpdate();
        };

        AppStatus.GetURLParams();
        AppStatus.OnStatusUpdate();
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

        if (page && (["home", "search", "random", "albums", "upload", "adv-search"].includes(page))) {
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

        const searchParams = getParameterByName("sp");

        if (searchParams) {
            AppStatus.SearchParams = searchParams;
        } else {
            AppStatus.SearchParams = "";
        }

        const split = getParameterByName("split");

        AppStatus.ListSplitMode = (split === "yes");
    }

    public static UpdateLayout() {
        if (AppStatus.CurrentPage === "search" && !AppStatus.CurrentSearch) {
            AppStatus.CurrentPage = "home";
        }

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

    public static IsPlayerVisible(): boolean {
        switch (AppStatus.CurrentLayout) {
        case "album":
        case "media-split":
        case "media":
            return true;
        default:
            return false;
        }
    }

    public static IsPageVisible(): boolean {
        switch (AppStatus.CurrentLayout) {
        case "initial":
        case "media-split":
            return true;
        default:
            return false;
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

        if (AppStatus.SearchParams) {
            params["sp"] = AppStatus.SearchParams;
        }

        if (AppStatus.ListSplitMode) {
            params["split"] = "yes";
        }

        return window.location.protocol + "//" + window.location.host + window.location.pathname + GenerateURIQuery(params);
    }

    public static UpdateURL(replaceState?: boolean) {
        if (AppStatus.CurrentAlbum >= 0 && AppStatus.CurrentMedia < 0) {
            if (AlbumsController.Loading) {
                return;
            }
            if (AlbumsController.CurrentAlbum !== AppStatus.CurrentAlbum) {
                return;
            }
            if (AlbumsController.CurrentAlbumData && AlbumsController.CurrentAlbumData.list.length > 0) {
                return;
            }
        }

        if (history.pushState) {
            const newURL = AppStatus.GetCurrentURL();
            // Update URL
            if (newURL !== location.href) {
                if (replaceState) {
                    window.history.replaceState({ path: newURL }, '', newURL);
                } else {
                    window.history.pushState({ path: newURL }, '', newURL);
                }
            }
        }
    }

    public static OnStatusUpdate(replaceState?: boolean) {
        AppStatus.UpdateLayout();

        AppEvents.Emit("app-status-update", AppStatus);

        AppStatus.UpdateURL(replaceState);
    }

    public static GoToPage(page: string) {
        AppStatus.CurrentPage = page;

        AppStatus.CurrentAlbum = -1;

        if (AppStatus.CurrentMedia >= 0) {
            AppStatus.ListSplitMode = true;
        }

        AppStatus.SearchParams = "";

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "right";

        AppStatus.OnStatusUpdate();
    }

    public static GoToPageNoSplit(page: string) {
        AppStatus.CurrentPage = page;

        AppStatus.CurrentAlbum = -1;
        AppStatus.CurrentMedia = -1;
        AppStatus.ListSplitMode = false;
        AppStatus.SearchParams = "";

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "right";

        AppStatus.OnStatusUpdate();
    }

    public static ExpandPage() {
        AppStatus.CurrentAlbum = -1;
        AppStatus.CurrentMedia = -1;
        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "right";

        AppStatus.OnStatusUpdate();
    }

    public static OnDeleteMedia() {
        AppStatus.CurrentMedia = -1;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "right";

        AppEvents.Emit("media-delete");

        AppStatus.OnStatusUpdate();
    }

    public static GoToSearch(search: string) {
        AppStatus.CurrentSearch = search;

        if (AppStatus.CurrentSearch) {
            if (AppStatus.CurrentPage !== "random") {
                AppStatus.CurrentPage = "search";
            }

            AppStatus.CurrentAlbum = -1;

            if (AppStatus.CurrentMedia >= 0) {
                AppStatus.ListSplitMode = true;
            }

            AppStatus.SearchParams = "";

            AppStatus.UpdateLayout();

            AppStatus.CurrentFocus = "right";
        }

        AppStatus.OnStatusUpdate();
    }

    public static ClearSearch() {
        AppStatus.CurrentSearch = "";

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "right";

        AppStatus.OnStatusUpdate();
    }

    public static ClickOnMedia(mediaId: number, split: boolean) {
        AppStatus.CurrentMedia = mediaId;

        if (split) {
            AppStatus.ListSplitMode = true;
        }

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate();
    }

    public static ClickOnAlbum(albumId: number) {
        AppStatus.CurrentAlbum = albumId;
        AppStatus.CurrentMedia = -1;
        AppStatus.CurrentSearch = "";

        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate();
    }

    public static ClickOnAlbumByMedia(albumId: number, mediaId: number) {
        AppStatus.CurrentAlbum = albumId;
        AppStatus.CurrentMedia = mediaId;
        AppStatus.CurrentSearch = "";

        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate();
    }

    public static ClickOnAlbumWithList(albumId: number, list: number[]) {
        AppStatus.CurrentAlbum = albumId;

        const pos = PlayerPreferences.GetAlbumPos(albumId);

        if (pos < list.length) {
            AppStatus.CurrentMedia = list[pos];
        } else if (list.length > 0) {
            AppStatus.CurrentMedia = list[0];
        } else {
            AppStatus.CurrentMedia = -1;
        }

        AppStatus.ListSplitMode = false;

        AppStatus.UpdateLayout();

        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate(true);
    }

    public static FocusLeft() {
        AppStatus.CurrentFocus = "left";

        AppStatus.OnStatusUpdate();
    }

    public static FocusRight() {
        AppStatus.CurrentFocus = "right";

        AppStatus.OnStatusUpdate();
    }

    public static ClosePage() {
        AppStatus.CurrentFocus = "left";
        AppStatus.ListSplitMode = false;
        AppStatus.UpdateLayout();
        AppStatus.OnStatusUpdate();
    }

    public static CloseAlbum() {
        AppStatus.CurrentFocus = "left";
        AppStatus.CurrentAlbum = -1;
        AppStatus.UpdateLayout();
        AppStatus.OnStatusUpdate();
    }

    public static ChangeSearchParams(params: string) {
        AppStatus.SearchParams = params;

        AppStatus.OnStatusUpdate();
    }

    public static PackSearchParams(page: number, pageSize: number, order: string): string {
        if (page === 0 && pageSize === 25 && order === "desc") {
            return "";
        }
        return page + "," + pageSize + "," + order;
    }

    public static UnPackSearchParams(params: string): { page: number, pageSize: number, order: string } {
        const res = {
            page: 0,
            pageSize: 25,
            order: "desc",
        };

        if (params) {
            const spl = params.split(",");
            res.page = parseInt(spl[0], 10) || 0;
            if (res.page < 0) {
                res.page = 0;
            }

            res.pageSize = parseInt(spl[1], 10) || 0;
            if (res.pageSize <= 0) {
                res.pageSize = 25;
            }

            res.order = spl[2] || "desc";

            if (res.order !== "desc" && res.order !== "asc") {
                res.order = "desc";
            }
        }

        return res;
    }
}
