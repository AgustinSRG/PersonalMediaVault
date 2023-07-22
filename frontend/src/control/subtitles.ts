// Subtitles controller

import { GetAssetURL, Request } from "@/utils/request";
import { findSubtitlesEntry, parseSRT, SubtitlesEntry } from "@/utils/srt";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { MediaController } from "./media";
import { PlayerPreferences } from "./player-preferences";

export class SubtitlesController {
    public static MediaId = -1;
    public static SelectedSubtitles = "";
    public static SubtitlesFileURL = "";
    public static Subtitles: SubtitlesEntry[] = [];

    public static Initialize() {
        AppEvents.AddEventListener("auth-status-changed", SubtitlesController.Load);
        AppEvents.AddEventListener("app-status-update", SubtitlesController.OnMediaChanged);
        AppEvents.AddEventListener("current-media-update", SubtitlesController.Load);

        SubtitlesController.MediaId = AppStatus.CurrentMedia;

        SubtitlesController.Load();
    }

    public static OnMediaChanged() {
        if (SubtitlesController.MediaId !== AppStatus.CurrentMedia) {
            SubtitlesController.MediaId = AppStatus.CurrentMedia;
            SubtitlesController.SelectedSubtitles = "";
            SubtitlesController.SubtitlesFileURL = "";
            SubtitlesController.Subtitles = [];
            SubtitlesController.Load();
        }
    }

    public static Load() {
        if (!MediaController.MediaData) {
            SubtitlesController.SelectedSubtitles = "";
            SubtitlesController.SubtitlesFileURL = "";
            SubtitlesController.Subtitles = [];
            AppEvents.Emit("subtitles-update");
            return;
        }

        const subtitles = MediaController.MediaData.subtitles || [];
        const prefSubtitles = PlayerPreferences.SelectedSubtitles;

        SubtitlesController.SelectedSubtitles = "";
        SubtitlesController.SubtitlesFileURL = "";
        SubtitlesController.Subtitles = [];

        for (const sub of subtitles) {
            if (sub.id === prefSubtitles) {
                SubtitlesController.SelectedSubtitles = sub.id;
                SubtitlesController.SubtitlesFileURL = GetAssetURL(sub.url);
                break;
            }
        }

        if (!SubtitlesController.SubtitlesFileURL) {
            AppEvents.Emit("subtitles-update");
            return;
        }

        Timeouts.Abort("subtitles-load");
        Request.Pending("subtitles-load", {
            method: "GET",
            url: SubtitlesController.SubtitlesFileURL,
        })
            .onSuccess((srtText) => {
                SubtitlesController.Subtitles = parseSRT(srtText);
                AppEvents.Emit("subtitles-update");
            })
            .onRequestError((err) => {
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit("unauthorized", false);
                    })
                    .add(404, "*", () => {
                        SubtitlesController.Subtitles = [];
                        AppEvents.Emit("subtitles-update");
                    })
                    .add("*", "*", () => {
                        // Retry
                        Timeouts.Set("subtitles-load", 1500, SubtitlesController.Load);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                console.error(err);
                // Retry
                Timeouts.Set("subtitles-load", 1500, SubtitlesController.Load);
            });
    }

    public static OnSubtitlesChanged() {
        if (SubtitlesController.SelectedSubtitles !== PlayerPreferences.SelectedSubtitles) {
            SubtitlesController.Load();
        }
    }

    public static GetSubtitlesLine(time: number): SubtitlesEntry {
        return findSubtitlesEntry(SubtitlesController.Subtitles, time);
    }
}

SubtitlesController.Initialize();
