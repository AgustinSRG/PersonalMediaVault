// Current media data controller

import { MediaAPI } from "@/api/api-media";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AlbumsController } from "./albums";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { AuthController } from "./auth";

export interface MediaEntry {
    id: number;
    type: number;
    title: string;
    description: string;
    tags: number[];
    thumbnail: string;
    duration: number;
}

export interface MediaData {
    id: number;

    type: number;

    title: string;
    description: string;
    tags: number[];

    thumbnail: string;

    duration: number;
    width: number;
    height: number;
    fps: number;

    ready: boolean;
    encoded: boolean;

    task: number;

    url: string;

    video_previews: string;
    video_previews_interval: number;

    resolutions: {
        width: number;
        height: number;
        fps: number;

        ready: boolean;
        task: number;

        url: string;
    }[],

    subtitles: {
        id: string;
        name: string;

        url: string;
    }[],

    audios: {
        id: string;
        name: string;

        url: string;
    }[],

    time_slices: {
        time: number,
        name: string,
    }[],

    force_start_beginning: boolean,

    img_notes: boolean,
    img_notes_url: string,

    ext_desc_url: string,
}

export class MediaController {
    public static Loading = true;
    public static MediaId = -1;
    public static MediaData: MediaData = null;

    public static Initialize() {
        AppEvents.AddEventListener("auth-status-changed", MediaController.Load);
        AppEvents.AddEventListener("app-status-update", MediaController.OnMediaChanged);

        MediaController.MediaId = AppStatus.CurrentMedia;

        MediaController.Load();
    }

    public static OnMediaChanged() {
        if (MediaController.MediaId !== AppStatus.CurrentMedia) {
            MediaController.MediaId = AppStatus.CurrentMedia;
            MediaController.Load();
        }
    }

    public static Load() {
        if (MediaController.MediaId < 0) {
            Timeouts.Abort("media-current-load");
            Request.Abort("media-current-load");

            MediaController.MediaData = null;
            AppEvents.Emit("current-media-update", null);
            MediaController.Loading = false;
            AppEvents.Emit("current-media-loading", false);

            return;
        }

        MediaController.MediaData = null;
        AppEvents.Emit("current-media-update", null);

        MediaController.Loading = true;
        AppEvents.Emit("current-media-loading", true);

        if (AuthController.Locked) {
            return; // Vault is locked
        }

        Timeouts.Abort("media-current-load");
        Request.Abort("media-current-load");

        if (AlbumsController.CheckAlbumNextPrefetch()) {
            return; // Pre-fetch
        }

        Request.Pending("media-current-load", MediaAPI.GetMedia(MediaController.MediaId)).onSuccess(media => {
            MediaController.MediaData = media;
            AppEvents.Emit("current-media-update", MediaController.MediaData);

            MediaController.Loading = false;
            AppEvents.Emit("current-media-loading", false);
        }).onRequestError(err => {
            Request.ErrorHandler()
                .add(401, "*", () => {
                    AppEvents.Emit("unauthorized", false);
                })
                .add(404, "*", () => {
                    MediaController.MediaData = null;
                    AppEvents.Emit("current-media-update", MediaController.MediaData);

                    MediaController.Loading = false;
                    AppEvents.Emit("current-media-loading", false);
                })
                .add("*", "*", () => {
                    // Retry
                    Timeouts.Set("media-current-load", 1500, MediaController.Load);
                })
                .handle(err);
        }).onUnexpectedError(err => {
            console.error(err);
            // Retry
            Timeouts.Set("media-current-load", 1500, MediaController.Load);
        });
    }

    public static NextPendingId = 0;

    public static GetPendingId() {
        MediaController.NextPendingId++;
        return "pending-check-" + Date.now() + "-" + MediaController.NextPendingId;
    }
}
