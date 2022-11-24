// Subtitles controller

import { GetAssetURL, Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { AppEvents } from "./app-events";
import { AppStatus } from "./app-status";
import { MediaController } from "./media";

export interface ImageNote {
    x: number;
    y: number;
    w: number;
    h: number;
    text: string;
}

function parseImageNotes(json: string): ImageNote[] {
    let o: any;
    try {
        o = JSON.parse(json);
    } catch (ex) {
        return [];
    }
    if (o && Array.isArray(o)) {
        return o.map(note => {
            if (note && typeof note === "object") {
                return {
                    x: parseInt(note.x, 10) || 0,
                    y: parseInt(note.y, 10) || 0,
                    w: parseInt(note.w, 10) || 0,
                    h: parseInt(note.h, 10) || 0,
                    text: (note.text || "") + "",
                };
            } else {
                return {
                    x: 0,
                    y: 0,
                    w: 0,
                    h: 0,
                    text: "",
                };
            }
        });
    } else {
        return [];
    }
}

export class ImageNotesController {
    public static MediaId = -1;
    public static NotesFileURL = "";
    public static Notes: ImageNote[] = [];

    public static Initialize() {
        AppEvents.AddEventListener("auth-status-changed", ImageNotesController.Load);
        AppEvents.AddEventListener("app-status-update", ImageNotesController.OnMediaChanged);
        AppEvents.AddEventListener("current-media-update", ImageNotesController.Load);

        ImageNotesController.MediaId = AppStatus.CurrentMedia;

        ImageNotesController.Load();
    }

    public static OnMediaChanged() {
        if (ImageNotesController.MediaId !== AppStatus.CurrentMedia) {
            ImageNotesController.MediaId = AppStatus.CurrentMedia;
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.Notes = [];
            ImageNotesController.Load();
        }
    }

    public static Load() {
        if (!MediaController.MediaData) {
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.Notes = [];
            AppEvents.Emit("img-notes-update");
            return;
        }

        if (!MediaController.MediaData.img_notes || !MediaController.MediaData.img_notes_url) {
            ImageNotesController.NotesFileURL = "";
            ImageNotesController.Notes = [];
            AppEvents.Emit("img-notes-update");
            return;
        }

        ImageNotesController.NotesFileURL = GetAssetURL(MediaController.MediaData.img_notes_url);
        ImageNotesController.Notes = [];

        Timeouts.Abort("img-notes-load");
        Request.Pending("img-notes-load",{
            method: "GET",
            url:  ImageNotesController.NotesFileURL,
        }).onSuccess(jsonNotes => {
            ImageNotesController.Notes = parseImageNotes(jsonNotes);
            AppEvents.Emit("img-notes-update");
        }).onRequestError(err => {
            Request.ErrorHandler()
                .add(401, "*", () => {
                    AppEvents.Emit("unauthorized", false);
                })
                .add(404, "*", () => {
                    ImageNotesController.Notes = [];
                    AppEvents.Emit("img-notes-update");
                })
                .add("*", "*", () => {
                    // Retry
                    Timeouts.Set("img-notes-load", 1500, ImageNotesController.Load);
                })
                .handle(err);
        }).onUnexpectedError(err => {
            console.error(err);
            // Retry
            Timeouts.Set("img-notes-load", 1500, ImageNotesController.Load);
        });
    }
}
