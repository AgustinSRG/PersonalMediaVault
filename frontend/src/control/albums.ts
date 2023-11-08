// Albums data controller

"use strict";

import { AlbumsAPI } from "@/api/api-albums";
import { MediaAPI } from "@/api/api-media";
import { Request } from "@/utils/request";
import { shuffleArray } from "@/utils/shuffle";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "./auth";
import { MediaController } from "./media";
import { setCachedAlbumPosition } from "./player-preferences";
import { Album, AlbumListItemMin, MediaData, MediaListItem } from "@/api/models";

export class AlbumsController {
    public static Albums: { [id: string]: AlbumListItemMin } = Object.create(null);

    public static Loading = true;
    public static InitiallyLoaded = false;

    public static Initialize() {
        AuthController.AddChangeEventListener(AlbumsController.Load);
        AuthController.AddChangeEventListener(AlbumsController.LoadCurrentAlbum);

        AppStatus.AddEventListener(AlbumsController.OnCurrentAlbumChanged);

        AlbumsController.CurrentAlbum = AppStatus.CurrentAlbum;

        AlbumsController.Load();
        AlbumsController.LoadCurrentAlbum();
    }

    public static Load() {
        AlbumsController.Loading = true;
        AppEvents.Emit("albums-loading", true);

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        Timeouts.Abort("albums-load");
        Request.Pending("albums-load", AlbumsAPI.GetAlbumsMin())
            .onSuccess((albums) => {
                AlbumsController.Albums = Object.create(null);

                for (const album of albums) {
                    AlbumsController.Albums[album.id + ""] = album;
                }

                AppEvents.Emit("albums-update", AlbumsController.Albums);
                AlbumsController.Loading = false;
                AppEvents.Emit("albums-loading", false);
                AlbumsController.InitiallyLoaded = true;
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set("albums-load", 1500, AlbumsController.Load);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                Timeouts.Set("albums-load", 1500, AlbumsController.Load);
            });
    }

    public static CurrentAlbum = -1;
    public static CurrentAlbumLoading = false;
    public static CurrentAlbumData: Album = null;

    public static OnCurrentAlbumChanged() {
        if (AppStatus.CurrentAlbum !== AlbumsController.CurrentAlbum) {
            AlbumsController.CurrentAlbum = AppStatus.CurrentAlbum;
            AlbumsController.CurrentAlbumData = null;
            AlbumsController.LoadCurrentAlbum();
        }
        AlbumsController.UpdateAlbumCurrentPos();
    }

    public static LoadCurrentAlbum() {
        if (AlbumsController.CurrentAlbum < 0) {
            Timeouts.Abort("album-current-load");
            Request.Abort("album-current-load");

            AlbumsController.CurrentAlbumData = null;
            AppEvents.Emit("current-album-update", null);
            AlbumsController.CurrentAlbumLoading = false;
            AppEvents.Emit("current-album-loading", false);

            AlbumsController.UpdateAlbumCurrentPos();

            return;
        }

        AlbumsController.CurrentAlbumLoading = true;
        AppEvents.Emit("current-album-loading", true);

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        Timeouts.Abort("album-current-load");
        Request.Pending("album-current-load", AlbumsAPI.GetAlbum(AlbumsController.CurrentAlbum))
            .onSuccess((album) => {
                AlbumsController.CurrentAlbumData = album;
                AppEvents.Emit("current-album-update", AlbumsController.CurrentAlbumData);

                AlbumsController.CurrentAlbumLoading = false;
                AppEvents.Emit("current-album-loading", false);

                AlbumsController.UpdateAlbumCurrentPos();
                AppStatus.UpdateURL();
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(404, "*", () => {
                        AlbumsController.CurrentAlbumData = null;
                        AppEvents.Emit("current-album-update", AlbumsController.CurrentAlbumData);

                        AlbumsController.CurrentAlbumLoading = false;
                        AppEvents.Emit("current-album-loading", false);

                        AppStatus.CloseAlbum();
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set("album-current-load", 1500, AlbumsController.LoadCurrentAlbum);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                Timeouts.Set("album-current-load", 1500, AlbumsController.LoadCurrentAlbum);
            });
    }

    public static GetAlbumsListCopy(): { id: number; name: string; nameLowerCase: string }[] {
        const result = [];

        for (const album of Object.values(AlbumsController.Albums)) {
            result.push({
                id: album.id,
                name: album.name,
                nameLowerCase: album.name.toLowerCase(),
            });
        }

        return result;
    }

    public static FindDuplicatedName(name: string): boolean {
        const nameLower = name.toLowerCase();

        for (const album of Object.values(AlbumsController.Albums)) {
            if (nameLower === album.name.toLowerCase()) {
                return true;
            }
        }

        return false;
    }

    public static OnChangedAlbum(albumId: number, noUpdateList?: boolean) {
        if (AlbumsController.CurrentAlbum === albumId) {
            AlbumsController.LoadCurrentAlbum();
        }
        AppEvents.Emit("albums-list-change");
        if (!noUpdateList) {
            AlbumsController.Load();
        }
    }

    public static MoveCurrentAlbumOrder(oldIndex: number, newIndex: number) {
        if (!AlbumsController.CurrentAlbumData) {
            return;
        }

        if (oldIndex < 0 || oldIndex >= AlbumsController.CurrentAlbumData.list.length) {
            return;
        }

        const albumId = AlbumsController.CurrentAlbumData.id;
        const mediaId = AlbumsController.CurrentAlbumData.list[oldIndex].id;

        AlbumsController.CurrentAlbumData.list.splice(newIndex, 0, AlbumsController.CurrentAlbumData.list.splice(oldIndex, 1)[0]);

        AppEvents.Emit("current-album-update", AlbumsController.CurrentAlbumData);

        AlbumsController.UpdateAlbumCurrentPos();

        // Update in server

        Request.Do(AlbumsAPI.MoveMediaInAlbum(albumId, mediaId, newIndex))
            .onSuccess(() => {
                AppEvents.Emit("album-order-saved");
                AppEvents.Emit("albums-list-change");
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
            });
    }

    public static CurrentAlbumPos = -1;
    public static CurrentPrev: MediaListItem = null;
    public static CurrentNext: MediaListItem = null;

    public static AlbumLoop = false;
    public static AlbumRandom = false;

    public static ToggleLoop() {
        AlbumsController.AlbumLoop = !AlbumsController.AlbumLoop;
        this.UpdateAlbumCurrentPos();
    }

    public static ToggleRandom() {
        AlbumsController.AlbumRandom = !AlbumsController.AlbumRandom;
        this.UpdateAlbumCurrentPos();
    }

    public static UpdateAlbumCurrentPos() {
        const mediaId = AppStatus.CurrentMedia;

        if (!AlbumsController.CurrentAlbumData || AlbumsController.CurrentAlbumLoading) {
            AlbumsController.CurrentAlbumPos = -1;
            AlbumsController.CurrentPrev = null;
            AlbumsController.CurrentNext = null;
            AppEvents.Emit("album-pos-update");
            AlbumsController.PreFetchAlbumNext();
            return;
        }

        if (mediaId < 0 && AlbumsController.CurrentAlbumData.list.length > 0) {
            const albumList = AlbumsController.CurrentAlbumData.list.map((a) => {
                return a.id;
            });
            AppStatus.ClickOnAlbumWithList(AlbumsController.CurrentAlbumData.id, albumList);
            return;
        }

        let mediaPos = -1;

        for (let i = 0; i < AlbumsController.CurrentAlbumData.list.length; i++) {
            if (mediaId === AlbumsController.CurrentAlbumData.list[i].id) {
                mediaPos = i;
                break;
            }
        }

        AlbumsController.CurrentAlbumPos = mediaPos;

        if (mediaPos >= 0) {
            if (AlbumsController.AlbumRandom) {
                const shuffled = shuffleArray(AlbumsController.CurrentAlbumData.list).filter((a) => {
                    return a.id !== mediaId;
                });

                AlbumsController.CurrentPrev = null;
                AlbumsController.CurrentNext = shuffled[1] || null;

                if (AlbumsController.AlbumLoop) {
                    if (AlbumsController.CurrentNext === null) {
                        AlbumsController.CurrentNext = AlbumsController.CurrentAlbumData.list[0] || null;
                    }
                }
            } else {
                AlbumsController.CurrentPrev = AlbumsController.CurrentAlbumData.list[mediaPos - 1] || null;
                AlbumsController.CurrentNext = AlbumsController.CurrentAlbumData.list[mediaPos + 1] || null;

                if (AlbumsController.AlbumLoop) {
                    if (AlbumsController.CurrentPrev === null) {
                        AlbumsController.CurrentPrev =
                            AlbumsController.CurrentAlbumData.list[AlbumsController.CurrentAlbumData.list.length - 1] || null;
                    }

                    if (AlbumsController.CurrentNext === null) {
                        AlbumsController.CurrentNext = AlbumsController.CurrentAlbumData.list[0] || null;
                    }
                }
            }
        } else {
            AlbumsController.CurrentPrev = null;
            AlbumsController.CurrentNext = null;
        }

        setCachedAlbumPosition(AlbumsController.CurrentAlbumData.id, AlbumsController.CurrentAlbumPos);

        AppEvents.Emit("album-pos-update");
        AlbumsController.PreFetchAlbumNext();
    }

    public static AvailableNextPrefetch = false;
    public static LoadingNext = false;
    public static NextMediaData: MediaData = null;

    public static PreFetchAlbumNext() {
        if (AlbumsController.CurrentNext === null || AlbumsController.CurrentNext.id === MediaController.MediaId) {
            Timeouts.Abort("album-next-prefetch-load");
            Request.Abort("album-next-prefetch-load");

            AlbumsController.NextMediaData = null;
            AlbumsController.LoadingNext = false;
            AlbumsController.AvailableNextPrefetch = false;
            AppEvents.Emit("album-next-prefetch");
            return;
        }

        AlbumsController.NextMediaData = null;
        AlbumsController.LoadingNext = true;
        AlbumsController.AvailableNextPrefetch = false;

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        const mediaId = AlbumsController.CurrentNext.id;

        Timeouts.Abort("album-next-prefetch-load");
        Request.Pending("album-next-prefetch-load", MediaAPI.GetMedia(mediaId))
            .onSuccess((media) => {
                AlbumsController.NextMediaData = media;
                AlbumsController.LoadingNext = false;
                AlbumsController.AvailableNextPrefetch = true;
                AppEvents.Emit("album-next-prefetch", mediaId);
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(404, "*", () => {
                        AlbumsController.NextMediaData = null;
                        AlbumsController.LoadingNext = false;
                        AlbumsController.AvailableNextPrefetch = true;
                        AppEvents.Emit("album-next-prefetch", mediaId);
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set("album-next-prefetch-load", 1500, AlbumsController.PreFetchAlbumNext);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                Timeouts.Set("album-next-prefetch-load", 1500, AlbumsController.PreFetchAlbumNext);
            });
    }

    public static CheckAlbumNextPrefetch(): boolean {
        if (AlbumsController.CurrentNext === null) {
            return false;
        }

        if (AlbumsController.CurrentNext.id !== MediaController.MediaId) {
            return false;
        }

        if (AlbumsController.AvailableNextPrefetch) {
            MediaController.SetMediaData(AlbumsController.NextMediaData);
            return true;
        } else {
            return false;
        }
    }

    public static HasPagePrev = false;
    public static HasPageNext = false;

    public static OnPageLoad(currentMediaIndex: number, pageSize: number, page: number, totalPages: number) {
        if (currentMediaIndex >= 0) {
            AlbumsController.HasPagePrev = currentMediaIndex > 0 || page > 0;
            AlbumsController.HasPageNext = currentMediaIndex < pageSize - 1 || page < totalPages - 1;
        } else {
            AlbumsController.HasPagePrev = false;
            AlbumsController.HasPageNext = false;
        }
        AppEvents.Emit("page-media-nav-update", AlbumsController.HasPagePrev, AlbumsController.HasPageNext);
    }

    public static OnPageUnload() {
        AlbumsController.HasPagePrev = false;
        AlbumsController.HasPageNext = false;
        AppEvents.Emit("page-media-nav-update", AlbumsController.HasPagePrev, AlbumsController.HasPageNext);
    }

    public static AlbumsPageSearch = "";
}
