// Albums data controller

import { AmbumsAPI } from "@/api/api-albums";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AuthController } from "./auth";

export interface AlbumEntry {
    id: number;
    name: string;
    list: number[];
}

export class AlbumsController {
    public static Abums: { [id: string]: AlbumEntry } = Object.create(null);

    public static Loading = true;

    public static Initailize() {
        AppEvents.AddEventListener("auth-status-changed", AlbumsController.Load);
        AlbumsController.Load();
    }

    public static Load() {
        AlbumsController.Loading = true;

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        AppEvents.Emit("albums-loading", true);
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
}
