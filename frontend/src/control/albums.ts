// Albums data controller

import { AmbumsAPI } from "@/api/api-albums";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { AuthController } from "./auth";
import { MediaEntry } from "./media";

export interface AlbumEntry {
    id: number;
    name: string;
    list: number[];
}

export interface AlbumData {
    id: number;
    name: string;
    list: MediaEntry[];
}

export class AlbumsController {
    public static Abums: { [id: string]: AlbumEntry } = Object.create(null);

    public static Loading = true;

    public static Initailize() {
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
        Request.Pending("albums-load", AmbumsAPI.GetAlbums()).onSuccess(albums => {
            AlbumsController.Abums = Object.create(null);

            for (const album of albums) {
                AlbumsController.Abums[album.id + ""] = album;
            }

            AppEvents.Emit("albums-update", AlbumsController.Abums);
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
            AlbumsController.LoadCurrentAlbum();
        }
    }

    public static LoadCurrentAlbum() {
        if (AlbumsController.CurrentAlbum < 0) {
            Timeouts.Abort("album-current-load");
            Request.Abort("album-current-load");

            AlbumsController.CurrentAlbumData = null;
            AppEvents.Emit("current-album-update", null);
            AlbumsController.CurrentAlbumLoading = false;
            AppEvents.Emit("current-album-loading", false);

            return;
        }

        AlbumsController.CurrentAlbumLoading = true;
        AppEvents.Emit("current-album-loading", true);

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        Timeouts.Abort("album-current-load");
        Request.Pending("album-current-load", AmbumsAPI.GetAlbum(AlbumsController.CurrentAlbum)).onSuccess(album => {
            AlbumsController.CurrentAlbumData = album;
            AppEvents.Emit("current-album-update", AlbumsController.CurrentAlbumData);

            AlbumsController.CurrentAlbumLoading = false;
            AppEvents.Emit("current-album-loading", false);
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
}
