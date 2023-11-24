// Upload controller

"use strict";

import { MediaAPI } from "@/api/api-media";
import { TagsAPI } from "@/api/api-tags";
import { Request } from "@asanrom/request-browser";
import { AlbumsController } from "./albums";
import { AppEvents } from "./app-events";
import { MediaController } from "./media";
import { TagsController } from "./tags";
import { UploadMediaAPI } from "@/api/api-media-upload";
import { EVENT_NAME_UNAUTHORIZED } from "./auth";
import { setLastUsedTag } from "./app-preferences";

const TICK_DELAY_MS = 500;

const REQUEST_PREFIX_UPLOAD = "upload-media-";
const REQUEST_PREFIX_CHECK_ENCRYPTION = "check-media-encryption-";

/**
 * Event triggered when the upload list updates
 * Handler like: (mode: "push" | "rm" | "update" | "clear", entry?: UploadEntryMin, index?: number) => void
 */
export const EVENT_NAME_UPLOAD_LIST_UPDATE = "upload-list-update";

/**
 * Removes extension if present, in order to get the title
 * @param fileName The file name
 * @returns The title
 */
function getTitleFromFileName(fileName: string): string {
    const parts = (fileName + "").split(".");
    if (parts.length > 1) {
        parts.pop();
        return parts.join(".");
    } else {
        return fileName;
    }
}

/**
 * Upload entry (basic)
 */
export interface UploadEntryMin {
    /**
     * Entry identifier
     */
    id: number;

    /**
     * File name
     */
    name: string;

    /**
     * File size (bytes)
     */
    size: number;

    /**
     * Upload status
     */
    status: "pending" | "uploading" | "encrypting" | "tag" | "ready" | "error";

    /**
     * Error (only relevant if status = 'error')
     */
    error: "" | "invalid-media" | "access-denied" | "deleted" | "server-error" | "no-internet";

    /**
     * Progress.
     * When status = 'uploading', progress = uploaded bytes percent in range [0, 100]
     * When status = 'encrypting', progress = encrypted bytes percent in range [0, 100]
     * When status = 'tag', progress = number of tags left
     */
    progress: number;

    /**
     * Media ID, only relevant if status = 'encrypting' | 'tag' | 'ready'
     */
    mid: number;
}

/**
 * Upload entry (full)
 */
interface UploadEntry extends UploadEntryMin {
    /**
     * File to upload
     */
    file: File;

    /**
     * True if busy (request in progress)
     */
    busy: boolean;

    /**
     * Last request timestamp (Unix milliseconds)
     */
    lastRequest: number;

    /**
     * ID of the album to add the media into. -1 means no album.
     */
    album: number;

    /**
     * List of remaining tags to add to the media
     */
    tags: string[];
}

/**
 * Management object to control uploads.
 */
export class UploadController {
    /**
     * Upload entries
     */
    public static Entries: UploadEntry[] = [];

    /**
     * Number of active uploads
     */
    private static UploadingCount = 0;

    /**
     * Max number of parallel active uploads
     */
    public static MaxParallelUploads = 1;

    /**
     * Counter to assign unique IDs to each entry
     */
    private static NextId = 0;

    /**
     * Interval to check the entries
     */
    private static timer: number = null;

    /**
     * Gets the list of entries
     * @returns The list of entries
     */
    public static GetEntries(): UploadEntryMin[] {
        return UploadController.Entries.map((e) => {
            return {
                id: e.id,
                name: e.name,
                size: e.size,
                mid: e.mid,
                status: e.status,
                error: e.error,
                progress: e.progress,
            };
        });
    }

    /**
     * Checks entries and initiates uploads if necessary
     */
    private static tick() {
        for (let index = 0; index < UploadController.Entries.length; index++) {
            const pending = UploadController.Entries[index];
            if (pending.status === "pending") {
                if (UploadController.UploadingCount < UploadController.MaxParallelUploads) {
                    UploadController.UploadMedia(pending, index);
                }
            } else if (pending.status === "encrypting") {
                if (!pending.busy && Date.now() - pending.lastRequest > 1000) {
                    UploadController.CheckEncryptionStatus(pending, index);
                }
            } else if (pending.status === "tag") {
                if (!pending.busy) {
                    UploadController.AddNextTag(pending, index);
                }
            }
        }
    }

    /**
     * Adds a file, creating a new entry
     * @param file The file to upload
     * @param album The album to add the media into. Set to -1 for no album.
     * @param tags The list of tags to add to the media.
     * @returns The created entry ID
     */
    public static AddFile(file: File, album: number, tags: string[]): number {
        UploadController.NextId++;
        const id = UploadController.NextId;
        UploadController.Entries.push({
            id: id,
            file: file,
            name: file.name,
            size: file.size,
            status: "pending",
            error: "",
            progress: 0,
            mid: -1,
            busy: false,
            lastRequest: 0,
            album: album,
            tags: tags.slice(),
        });
        if (UploadController.timer === null) {
            UploadController.timer = setInterval(UploadController.tick, TICK_DELAY_MS);
        }
        UploadController.Emit("push", {
            id: id,
            name: file.name,
            size: file.size,
            status: "pending",
            error: "",
            progress: 0,
            mid: -1,
        });
        return id;
    }

    /**
     * Retries uploading a file, when it resulted in error
     * @param id The entry ID
     */
    public static TryAgain(id: number) {
        for (let i = 0; i < UploadController.Entries.length; i++) {
            if (UploadController.Entries[i].id === id) {
                if (UploadController.Entries[i].status !== "error") {
                    return;
                }

                UploadController.Entries[i].error = "";
                UploadController.Entries[i].status = "pending";
                UploadController.Entries[i].progress = 0;

                UploadController.CheckEmptyList();
                UploadController.Emit("update", UploadController.Entries[i], i);
                return;
            }
        }
    }

    /**
     * Checks if the list if empty.
     * Meaning there are no pending uploads.
     */
    public static CheckEmptyList() {
        let isEmpty = false;
        for (const entry of UploadController.Entries) {
            if (entry.status !== "ready" && entry.status !== "error") {
                isEmpty = false;
                break;
            }
        }
        if (isEmpty) {
            if (UploadController.timer !== null) {
                clearInterval(UploadController.timer);
                UploadController.timer = null;
            }
        } else {
            if (UploadController.timer === null) {
                UploadController.timer = setInterval(UploadController.tick, TICK_DELAY_MS);
            }
        }
    }

    /**
     * Removes an entry, aborting any requests.
     * @param id The entry ID
     */
    public static RemoveFile(id: number) {
        // Abort requests
        Request.Abort(REQUEST_PREFIX_UPLOAD + id);
        Request.Abort(REQUEST_PREFIX_CHECK_ENCRYPTION + id);

        // Remove from the array
        for (let i = 0; i < UploadController.Entries.length; i++) {
            if (UploadController.Entries[i].id === id) {
                if (UploadController.Entries[i].status === "encrypting") {
                    UploadController.UploadingCount--;
                }

                UploadController.Entries.splice(i, 1);
                UploadController.CheckEmptyList();
                UploadController.Emit("rm", null, i);
                return;
            }
        }
    }

    /**
     * Clears any completed uploads from the list
     */
    public static ClearList() {
        const entries = UploadController.Entries.slice();

        for (const entry of entries) {
            if (entry.status === "ready" || entry.status === "error") {
                UploadController.RemoveFile(entry.id);
            }
        }

        UploadController.CheckEmptyList();
        UploadController.Emit("clear");
    }

    /**
     * Cancels all pending uploads
     */
    public static CancelAll() {
        const entries = UploadController.Entries.slice();

        for (const entry of entries) {
            UploadController.RemoveFile(entry.id);
        }

        UploadController.CheckEmptyList();
        UploadController.Emit("clear");
    }

    /**
     * Uploads the media file
     * @param m The entry
     * @param index The index of the entry in the array
     */
    private static UploadMedia(m: UploadEntry, index: number) {
        UploadController.UploadingCount++;

        m.status = "uploading";
        m.progress = 0;
        UploadController.Emit("update", m, index);

        Request.Pending(REQUEST_PREFIX_UPLOAD + m.id, UploadMediaAPI.UploadMedia(getTitleFromFileName(m.name), m.file, m.album))
            .onUploadProgress((loaded, total) => {
                m.progress = Math.round(((loaded * 100) / total) * 100) / 100;
                UploadController.Emit("update", m, index);
            })
            .onSuccess((data) => {
                m.mid = data.media_id;
                m.status = "encrypting";
                m.progress = 0;
                UploadController.Emit("update", m, index);
            })
            .onCancel(() => {
                UploadController.UploadingCount--;
            })
            .onRequestError((err) => {
                UploadController.UploadingCount--;
                Request.ErrorHandler()
                    .add(400, "*", () => {
                        m.error = "invalid-media";
                        m.status = "error";
                    })
                    .add(401, "*", () => {
                        m.error = "access-denied";
                        m.status = "error";
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(403, "*", () => {
                        m.error = "access-denied";
                        m.status = "error";
                    })
                    .add(500, "*", () => {
                        m.error = "server-error";
                        m.status = "error";
                    })
                    .add("*", "*", () => {
                        m.error = "no-internet";
                        m.status = "error";
                    })
                    .handle(err);
                UploadController.CheckEmptyList();
                UploadController.Emit("update", m, index);
            })
            .onUnexpectedError((err) => {
                UploadController.UploadingCount--;
                m.error = "no-internet";
                console.error(err);
                m.status = "error";
                UploadController.CheckEmptyList();
                UploadController.Emit("update", m, index);
            });
    }

    /**
     * Checks encryption status of the media
     * @param m The entry
     * @param index The index of the entry in the array
     */
    private static CheckEncryptionStatus(m: UploadEntry, index: number) {
        if (m.busy) {
            return;
        }

        m.busy = true;
        m.lastRequest = Date.now();

        Request.Pending(REQUEST_PREFIX_CHECK_ENCRYPTION + m.id, MediaAPI.GetMedia(m.mid))
            .onSuccess((media) => {
                m.busy = false;
                if (media.ready) {
                    if (m.tags.length > 0) {
                        m.status = "tag";
                        m.progress = m.tags.length;
                    } else {
                        m.status = "ready";
                    }

                    UploadController.Emit("update", m, index);

                    if (m.album !== -1) {
                        AlbumsController.OnChangedAlbum(m.album, true);
                    }

                    if (MediaController.MediaId === m.mid) {
                        MediaController.Load();
                    }

                    UploadController.UploadingCount--;
                } else {
                    m.progress = Math.max(m.progress, media.ready_p);
                    UploadController.Emit("update", m, index);
                }
            })
            .onCancel(() => {
                m.busy = false;
            })
            .onRequestError((err) => {
                m.busy = false;
                Request.ErrorHandler()
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(404, "*", () => {
                        m.error = "deleted";
                        m.status = "error";
                        UploadController.UploadingCount--;
                        UploadController.CheckEmptyList();
                        UploadController.Emit("update", m, index);
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                m.busy = false;
                console.error(err);
            });
    }

    /**
     * Adds next tag to the media
     * @param m The entry
     * @param index The index of the entry in the array
     */
    private static AddNextTag(m: UploadEntry, index: number) {
        if (m.busy) {
            return;
        }

        if (m.tags.length === 0) {
            m.status = "ready";
            UploadController.Emit("update", m, index);
            UploadController.CheckEmptyList();
            if (MediaController.MediaId === m.mid) {
                MediaController.Load();
            } else if (AlbumsController.CurrentNext && AlbumsController.CurrentNext.id === m.mid) {
                AlbumsController.PreFetchAlbumNext();
            }
            return;
        }

        m.busy = true;

        const tag = m.tags[0];
        const mediaId = m.mid;

        Request.Do(TagsAPI.TagMedia(mediaId, tag))
            .onSuccess((res) => {
                setLastUsedTag(res.id);
                m.tags.shift(); // Remove tag from list
                m.progress = m.tags.length;
                m.busy = false;
                UploadController.Emit("update", m, index);
                TagsController.AddTag(res.id, res.name);
            })
            .onCancel(() => {
                m.busy = false;
            })
            .onRequestError((err) => {
                m.busy = false;
                Request.ErrorHandler()
                    .add(400, "*", () => {
                        m.tags.shift(); // Invalid tag name
                        m.progress = m.tags.length;
                        UploadController.Emit("update", m, index);
                    })
                    .add(401, "*", () => {
                        AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                    })
                    .add(403, "*", () => {
                        m.error = "access-denied";
                        m.status = "error";
                        UploadController.Emit("update", m, index);
                        UploadController.CheckEmptyList();
                    })
                    .add(404, "*", () => {
                        m.error = "deleted";
                        m.status = "error";
                        UploadController.Emit("update", m, index);
                        UploadController.CheckEmptyList();
                    })
                    .handle(err);
            })
            .onUnexpectedError((err) => {
                m.busy = false;
                console.error(err);
            });
    }

    /**
     * Emits list update event
     * @param mode Mode
     * @param entry Entry
     * @param index Index
     */
    private static Emit(mode: "push" | "rm" | "update" | "clear", entry?: UploadEntryMin, index?: number) {
        AppEvents.Emit(EVENT_NAME_UPLOAD_LIST_UPDATE, mode, entry, index);
    }
}
