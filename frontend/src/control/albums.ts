// Albums data controller

import { AlbumsAPI } from "@/api/api-albums";
import { MediaAPI } from "@/api/api-media";
import { Request } from "@/utils/request";
import { shuffleArray } from "@/utils/shuffle";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { AuthController } from "./auth";
import { MediaController, MediaData, MediaEntry } from "./media";
import { PlayerPreferences } from "./player-preferences";

export interface AlbumEntry {
    id: number;
    name: string;
    size: number;
    thumbnail: string;
}

export interface AlbumEntryMin {
    id: number;
    name: string;
}

export interface AlbumData {
    id: number;
    name: string;
    list: MediaEntry[];
}

export class AlbumsController {
    public static Albums: { [id: string]: AlbumEntryMin } = Object.create(null);

    public static Loading = true;

    public static Initialize() {
        AppEvents.AddEventListener("auth-status-changed", AlbumsController.Load);
        AppEvents.AddEventListener("auth-status-changed", AlbumsController.LoadCurrentAlbum);

        AppEvents.AddEventListener("app-status-update", AlbumsController.OnCurrentAlbumChanged);

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
        Request.Pending("albums-load", AlbumsAPI.GetAlbumsMin()).onSuccess(albums => {
            AlbumsController.Albums = Object.create(null);

            for (const album of albums) {
                AlbumsController.Albums[album.id + ""] = album;
            }

            AppEvents.Emit("albums-update", AlbumsController.Albums);
            AlbumsController.Loading = false;
            AppEvents.Emit("albums-loading", false);
        }).onRequestError(err => {
            Request.ErrorHandler()
                .add(401, "*", () => {
                    AppEvents.Emit("unauthorized", false);
                })
                .add("*", "*", () => {
                    // Retry
                    Timeouts.Set("albums-load", 1500, AlbumsController.Load);
                })
                .handle(err);
        }).onUnexpectedError(err => {
            console.error(err);
            // Retry
            Timeouts.Set("albums-load", 1500, AlbumsController.Load);
        });
    }

    public static CurrentAlbum = -1;
    public static CurrentAlbumLoading = false;
    public static CurrentAlbumData: AlbumData = null;

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
        Request.Pending("album-current-load", AlbumsAPI.GetAlbum(AlbumsController.CurrentAlbum)).onSuccess(album => {
            AlbumsController.CurrentAlbumData = album;
            AppEvents.Emit("current-album-update", AlbumsController.CurrentAlbumData);

            AlbumsController.CurrentAlbumLoading = false;
            AppEvents.Emit("current-album-loading", false);

            AlbumsController.UpdateAlbumCurrentPos();
        }).onRequestError(err => {
            Request.ErrorHandler()
                .add(401, "*", () => {
                    AppEvents.Emit("unauthorized", false);
                })
                .add(404, "*", () => {
                    AlbumsController.CurrentAlbumData = null;
                    AppEvents.Emit("current-album-update", AlbumsController.CurrentAlbumData);

                    AlbumsController.CurrentAlbumLoading = false;
                    AppEvents.Emit("current-album-loading", false);

                    AppStatus.CloseAbum();
                })
                .add("*", "*", () => {
                    // Retry
                    Timeouts.Set("album-current-load", 1500, AlbumsController.LoadCurrentAlbum);
                })
                .handle(err);
        }).onUnexpectedError(err => {
            console.error(err);
            // Retry
            Timeouts.Set("album-current-load", 1500, AlbumsController.LoadCurrentAlbum);
        });
    }

    public static GetAlbumsListCopy(): { id: number, name: string, nameLowerCase: string }[] {
        const result = [];

        for (const album of Object.values(AlbumsController.Albums)) {
            result.push({
                id: album.id,
                name: album.name,
                nameLowerCase: album.name.toLowerCase(),
            })
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
        AlbumsController.CurrentAlbumData.list.splice(
            newIndex,
            0,
            AlbumsController.CurrentAlbumData.list.splice(oldIndex, 1)[0]
        );

        AppEvents.Emit("current-album-update", AlbumsController.CurrentAlbumData);

        AlbumsController.UpdateAlbumCurrentPos();

        // Update in server
        const albumId = AlbumsController.CurrentAlbumData.id;
        const albumList = AlbumsController.CurrentAlbumData.list.map(a => {
            return a.id;
        });
        Request.Do(AlbumsAPI.SetAlbumOrder(albumId, albumList))
            .onSuccess(() => {
                AppEvents.Emit("album-order-saved");
                AppEvents.Emit("albums-list-change");
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit("unauthorized");
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
            });
    }

    public static CurrentAlbumPos = -1;
    public static CurrentPrev: MediaEntry = null;
    public static CurrentNext: MediaEntry = null;

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
            const albumList = AlbumsController.CurrentAlbumData.list.map(a => {
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
                const shuffled = shuffleArray(AlbumsController.CurrentAlbumData.list).filter(a => {
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
                        AlbumsController.CurrentPrev = AlbumsController.CurrentAlbumData.list[AlbumsController.CurrentAlbumData.list.length - 1] || null;
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

        PlayerPreferences.SetAlbumPos(AlbumsController.CurrentAlbumData.id, AlbumsController.CurrentAlbumPos);

        AppEvents.Emit("album-pos-update");
        AlbumsController.PreFetchAlbumNext();
    }

    public static AvailableNextPrefetch = false;
    public static LoadingNext = false;
    public static NextMediaData: MediaData = null;
    public static LoadingNextWaiting = false;

    public static PreFetchAlbumNext() {
        if (AlbumsController.CurrentNext === null || AlbumsController.CurrentNext.id === MediaController.MediaId) {
            Timeouts.Abort("album-next-prefetch-load");
            Request.Abort("album-next-prefetch-load");

            AlbumsController.NextMediaData = null;
            AlbumsController.LoadingNext = false;
            AlbumsController.AvailableNextPrefetch = false;
            if (AlbumsController.LoadingNextWaiting) {
                AlbumsController.LoadingNextWaiting = false;
                MediaController.Load();
            }
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
        Request.Pending("album-next-prefetch-load", MediaAPI.GetMedia(mediaId)).onSuccess(media => {
            AlbumsController.NextMediaData = media;
            AlbumsController.LoadingNext = false;
            AlbumsController.AvailableNextPrefetch = true;
            AlbumsController.OnAlbumNextPrefetchDone(mediaId);
        }).onRequestError(err => {
            Request.ErrorHandler()
                .add(401, "*", () => {
                    AppEvents.Emit("unauthorized", false);
                })
                .add(404, "*", () => {
                    AlbumsController.NextMediaData = null;
                    AlbumsController.LoadingNext = false;
                    AlbumsController.AvailableNextPrefetch = true;
                    AlbumsController.OnAlbumNextPrefetchDone(mediaId);
                })
                .add("*", "*", () => {
                    // Retry
                    Timeouts.Set("album-next-prefetch-load", 1500, AlbumsController.PreFetchAlbumNext);
                })
                .handle(err);
        }).onUnexpectedError(err => {
            console.error(err);
            // Retry
            Timeouts.Set("album-next-prefetch-load", 1500, AlbumsController.PreFetchAlbumNext);
        });
    }

    public static OnAlbumNextPrefetchDone(mid: number) {
        AppEvents.Emit("album-next-prefetch");
        if (AlbumsController.LoadingNextWaiting && MediaController.MediaId === mid) {
            MediaController.MediaData = AlbumsController.NextMediaData;
            AppEvents.Emit("current-media-update", MediaController.MediaData);

            MediaController.Loading = false;
            AppEvents.Emit("current-media-loading", false);
        }
        AlbumsController.LoadingNextWaiting = false;
    }

    public static CheckAlbumNextPrefetch(): boolean {
        if (AlbumsController.CurrentNext === null) {
            return false;
        }

        if (AlbumsController.CurrentNext.id !== MediaController.MediaId) {
            return false;
        }

        if (AlbumsController.AvailableNextPrefetch) {
            MediaController.MediaData = AlbumsController.NextMediaData;
            AppEvents.Emit("current-media-update", MediaController.MediaData);

            MediaController.Loading = false;
            AppEvents.Emit("current-media-loading", false);
            return true;
        } else if (AlbumsController.LoadingNextWaiting) {
            AlbumsController.LoadingNextWaiting = true;
            return true;
        } else {
            return false;
        }
    }
}
