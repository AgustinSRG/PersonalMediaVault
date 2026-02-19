// Upload controller

"use strict";

import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { AlbumsController } from "./albums";
import { MediaController } from "./media";
import { TagsController } from "./tags";
import { getMaxParallelUploads, setLastUsedTag, setMaxParallelUploads } from "./app-preferences";
import { apiUploadMedia } from "@/api/api-media-upload";
import { apiMediaGetMedia } from "@/api/api-media";
import { apiTagsTagMedia } from "@/api/api-tags";
import { apiAlbumsAddMediaToAlbum } from "@/api/api-albums";
import { removeFromArray } from "@/utils/objects";
import {
    emitAppEvent,
    EVENT_NAME_UPLOAD_LIST_ENTRY_NEW,
    EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY,
    EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED,
    EVENT_NAME_UPLOAD_LIST_CLEAR,
    EVENT_NAME_UPLOAD_LIST_ENTRY_READY,
    EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR,
    EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS,
    EVENT_NAME_UNAUTHORIZED,
} from "./app-events";

/**
 * Max number of uploads in the completed lists (ready, error)
 * to keep before starting to remove them
 */
const MAX_COMPLETED_UPLOADS = 1024;

/**
 * Interval to check for encryption progress
 * (In milliseconds)
 */
const ENCRYPTION_CHECK_INTERVAL_MS = 250;

/**
 * Delay to wait after an error
 * (In milliseconds)
 */
const RETRY_DELAY_MS = 1500;

const REQUEST_PREFIX_UPLOAD = "upload-media-";
const REQUEST_PREFIX_CHECK_ENCRYPTION = "check-media-encryption-";

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
    status: "pending" | "uploading" | "encrypting" | "tag" | "album" | "ready" | "error";

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

    /**
     * Timer to retry an operation, after its failure
     */
    retryTimer: ReturnType<typeof setTimeout> | null;
}

/**
 * Management object to control uploads.
 */
export class UploadController {
    /**
     * Upload entries
     */
    private static Entries = {
        pending: [] as UploadEntry[],
        ready: [] as UploadEntry[],
        error: [] as UploadEntry[],
    };

    /**
     * Map of upload entries
     */
    private static EntriesMap: Map<number, UploadEntry> = new Map();

    /**
     * Upload queue
     */
    private static UploadQueue: UploadEntry[] = [];

    /**
     * Queue for albums
     */
    private static AlbumQueue: Map<number, UploadEntry[]> = new Map();

    /**
     * Number of pending uploads (not ready or error)
     */
    private static PendingCount = 0;

    /**
     * Number of active uploads
     */
    private static UploadingCount = 0;

    /**
     * Max number of parallel active uploads
     */
    private static MaxParallelUploads = getMaxParallelUploads();

    /**
     * Counter to assign unique IDs to each entry
     */
    private static NextId = 0;

    /**
     * Gets the list of pending entries
     * @returns The list of entries
     */
    public static GetPendingEntries(): UploadEntryMin[] {
        return UploadController.Entries.pending.map(minimizeEntry);
    }

    /**
     * Gets the list of ready entries
     * @returns The list of entries
     */
    public static GetReadyEntries(): UploadEntryMin[] {
        return UploadController.Entries.ready.map(minimizeEntry);
    }

    /**
     * Gets the list of ready entries
     * @returns The list of entries
     */
    public static GetErrorEntries(): UploadEntryMin[] {
        return UploadController.Entries.error.map(minimizeEntry);
    }

    /**
     * Gets the max parallel uploads
     * @returns The max parallel uploads
     */
    public static GetMaxParallelUploads(): number {
        return UploadController.MaxParallelUploads;
    }

    /**
     * Sets the max parallel uploads
     * @param maxParallelUploads The max parallel uploads
     */
    public static SetMaxParallelUploads(maxParallelUploads: number) {
        UploadController.MaxParallelUploads = maxParallelUploads;
        setMaxParallelUploads(maxParallelUploads);
        UploadController.CheckQueue();
    }

    /**
     * Checks the upload queue
     */
    private static CheckQueue() {
        while (UploadController.UploadQueue.length > 0 && UploadController.UploadingCount < UploadController.MaxParallelUploads) {
            UploadController.UploadMedia(UploadController.UploadQueue.shift());
        }

        UploadController.AlbumQueue.forEach((queue) => {
            if (queue.length === 0) {
                return;
            }

            const headEntry = queue[0];

            if (headEntry.busy || headEntry.status !== "album") {
                return;
            }

            UploadController.InsertIntoAlbum(headEntry);
        });
    }

    /**
     * Adds entry to album queue
     * @param entry The entry
     */
    private static AddToAlbumQueue(entry: UploadEntry) {
        if (entry.album < 0) {
            return;
        }

        if (UploadController.AlbumQueue.has(entry.album)) {
            const q = UploadController.AlbumQueue.get(entry.album);
            q.push(entry);
        } else {
            UploadController.AlbumQueue.set(entry.album, [entry]);
        }
    }

    /**
     * Removes entry from album queue
     * @param entry The upload entry
     */
    private static RemoveFromAlbumQueue(entry: UploadEntry) {
        if (entry.album < 0) {
            return;
        }

        if (UploadController.AlbumQueue.has(entry.album)) {
            const q = UploadController.AlbumQueue.get(entry.album);
            removeFromArray(q, entry);

            if (q.length === 0) {
                UploadController.AlbumQueue.delete(entry.album);
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

        const entry: UploadEntry = {
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
            retryTimer: null,
        };

        UploadController.EntriesMap.set(entry.id, entry);
        UploadController.Entries.pending.push(entry);
        UploadController.UploadQueue.push(entry);

        if (album >= 0) {
            if (UploadController.AlbumQueue.has(album)) {
                const q = UploadController.AlbumQueue.get(album);
                q.push(entry);
            } else {
                UploadController.AlbumQueue.set(album, [entry]);
            }
        }

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_NEW, minimizeEntry(entry));

        UploadController.CheckQueue();

        return id;
    }

    /**
     * Retries uploading a file, when it resulted in error
     * @param id The entry ID
     */
    public static TryAgain(id: number) {
        const entry = UploadController.EntriesMap.get(id);

        if (!entry || entry.status !== "error") {
            return;
        }

        entry.error = "";
        entry.status = "pending";
        entry.progress = 0;

        removeFromArray(UploadController.Entries.error, entry);
        UploadController.Entries.pending.push(entry);

        UploadController.UploadQueue.push(entry);

        UploadController.AddToAlbumQueue(entry);

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY, minimizeEntry(entry));
    }

    /**
     * Removes an entry, aborting any requests.
     * @param id The entry ID
     */
    public static RemoveFile(id: number) {
        const entry = UploadController.EntriesMap.get(id);

        if (!entry) {
            return;
        }

        UploadController.EntriesMap.delete(entry.id);

        // Abort requests
        abortNamedApiRequest(REQUEST_PREFIX_UPLOAD + id);
        abortNamedApiRequest(REQUEST_PREFIX_CHECK_ENCRYPTION + id);

        // Abort timers
        if (entry.retryTimer) {
            clearTimeout(entry.retryTimer);
        }

        // Remove from lists

        switch (entry.status) {
            case "ready":
                removeFromArray(UploadController.Entries.ready, entry);
                break;
            case "error":
                removeFromArray(UploadController.Entries.error, entry);
                break;
            default:
                removeFromArray(UploadController.Entries.pending, entry);
                removeFromArray(UploadController.UploadQueue, entry);
        }

        UploadController.RemoveFromAlbumQueue(entry);

        // Decrease the uploading count in case if was in the encrypting phase
        // In the uploading phase, since it is a single request, it is handled in the cancel callback

        if (entry.status === "encrypting") {
            UploadController.UploadingCount--;
        }

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED, minimizeEntry(entry));
    }

    /**
     * Clears any completed uploads from the list
     */
    public static ClearList() {
        const completedUploads = UploadController.Entries.ready.concat(UploadController.Entries.error);

        for (const entry of completedUploads) {
            UploadController.EntriesMap.delete(entry.id);
        }

        UploadController.Entries.ready = [];
        UploadController.Entries.error = [];

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_CLEAR);
    }

    /**
     * Cancels all pending uploads
     */
    public static CancelAll() {
        const pendingUploads = UploadController.Entries.pending;

        for (const entry of pendingUploads) {
            UploadController.RemoveFile(entry.id);
        }

        const completedUploads = UploadController.Entries.ready.concat(UploadController.Entries.error);

        for (const entry of completedUploads) {
            UploadController.EntriesMap.delete(entry.id);
        }

        UploadController.Entries.pending = [];
        UploadController.Entries.ready = [];
        UploadController.Entries.error = [];

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_CLEAR);
    }

    /**
     * Checks a completed list for its max length
     * @param list The completed list (ready or error)
     */
    private static CheckCompletedList(list: UploadEntry[]) {
        if (list.length > MAX_COMPLETED_UPLOADS) {
            const toRemove = list.shift();

            UploadController.EntriesMap.delete(toRemove.id);

            emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED, minimizeEntry(toRemove));
        }
    }

    /**
     * Call when entry is ready
     * @param m The upload entry
     */
    private static OnUploadEntryReady(m: UploadEntry) {
        removeFromArray(UploadController.Entries.pending, m);
        UploadController.Entries.ready.push(m);

        UploadController.RemoveFromAlbumQueue(m);

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_READY, minimizeEntry(m));

        UploadController.CheckQueue();

        if (MediaController.MediaId === m.mid) {
            MediaController.Load();
        }

        if (AlbumsController.CurrentNext && AlbumsController.CurrentNext.id === m.mid) {
            AlbumsController.PreFetchAlbumNext();
        }

        if (m.album !== -1) {
            AlbumsController.OnChangedAlbum(m.album, true);
        }

        UploadController.CheckCompletedList(UploadController.Entries.ready);
    }

    /**
     * Call when an entry reaches the error status
     * Moves the entry and emits the event
     * @param m The entry
     */
    private static OnEntryError(m: UploadEntry) {
        removeFromArray(UploadController.Entries.pending, m);
        UploadController.Entries.error.push(m);

        UploadController.RemoveFromAlbumQueue(m);

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR, minimizeEntry(m));

        UploadController.CheckQueue();

        UploadController.CheckCompletedList(UploadController.Entries.error);
    }

    /**
     * Uploads the media file
     * @param m The entry
     */
    private static UploadMedia(m: UploadEntry) {
        UploadController.UploadingCount++;

        m.status = "uploading";
        m.progress = 0;
        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

        makeNamedApiRequest(REQUEST_PREFIX_UPLOAD + m.id, apiUploadMedia(getTitleFromFileName(m.name), m.file, -1))
            .onUploadProgress((loaded, total) => {
                m.progress = Math.round(((loaded * 100) / total) * 100) / 100;
                emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));
            })
            .onSuccess((data) => {
                m.mid = data.media_id;
                m.status = "encrypting";
                m.progress = 0;

                emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                m.retryTimer = setTimeout(() => {
                    UploadController.CheckEncryptionStatus(m);
                }, ENCRYPTION_CHECK_INTERVAL_MS);
            })
            .onCancel(() => {
                UploadController.UploadingCount--;
                UploadController.RemoveFile(m.id);
                UploadController.CheckQueue();
            })
            .onRequestError((err, handleErr) => {
                UploadController.UploadingCount--;

                handleErr(err, {
                    unauthorized: () => {
                        m.error = "access-denied";
                        m.status = "error";
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    accessDenied: () => {
                        m.error = "access-denied";
                        m.status = "error";
                    },
                    invalidMediaFile: () => {
                        m.error = "invalid-media";
                        m.status = "error";
                    },
                    badRequest: () => {
                        m.error = "invalid-media";
                        m.status = "error";
                    },
                    serverError: () => {
                        m.error = "server-error";
                        m.status = "error";
                    },
                    networkError: () => {
                        m.error = "no-internet";
                        m.status = "error";
                    },
                });

                UploadController.OnEntryError(m);
            })
            .onUnexpectedError((err) => {
                UploadController.UploadingCount--;

                m.error = "no-internet";
                console.error(err);
                m.status = "error";

                UploadController.OnEntryError(m);
            });
    }

    /**
     * Checks encryption status of the media
     * @param m The entry
     */
    private static CheckEncryptionStatus(m: UploadEntry) {
        if (m.busy) {
            return;
        }

        m.retryTimer = null;

        m.busy = true;
        m.lastRequest = Date.now();

        makeNamedApiRequest(REQUEST_PREFIX_CHECK_ENCRYPTION + m.id, apiMediaGetMedia(m.mid))
            .onSuccess((media) => {
                m.busy = false;
                if (media.ready) {
                    if (m.tags.length > 0) {
                        m.status = "tag";
                        m.progress = m.tags.length;
                    } else if (m.album >= 0) {
                        m.status = "album";
                    } else {
                        m.status = "ready";
                    }

                    UploadController.UploadingCount--;

                    if (m.status === "ready") {
                        UploadController.OnUploadEntryReady(m);
                    } else {
                        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));
                        UploadController.CheckQueue();

                        if (m.status === "tag") {
                            UploadController.AddNextTag(m);
                        }
                    }
                } else {
                    m.progress = Math.max(m.progress, media.ready_p);
                    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                    m.retryTimer = setTimeout(() => {
                        UploadController.CheckEncryptionStatus(m);
                    }, ENCRYPTION_CHECK_INTERVAL_MS);
                }
            })
            .onCancel(() => {
                m.busy = false;
            })
            .onRequestError((err, handleErr) => {
                m.busy = false;
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
                    },
                    notFound: () => {
                        m.error = "deleted";
                        m.status = "error";

                        UploadController.UploadingCount--;

                        UploadController.OnEntryError(m);
                    },
                    temporalError: () => {
                        // Retry

                        m.retryTimer = setTimeout(() => {
                            UploadController.CheckEncryptionStatus(m);
                        }, RETRY_DELAY_MS);
                    },
                });
            })
            .onUnexpectedError((err) => {
                m.busy = false;
                console.error(err);

                // Retry

                m.retryTimer = setTimeout(() => {
                    UploadController.CheckEncryptionStatus(m);
                }, RETRY_DELAY_MS);
            });
    }

    /**
     * Adds next tag to the media
     * @param m The entry
     */
    private static AddNextTag(m: UploadEntry) {
        if (m.busy) {
            return;
        }

        m.retryTimer = null;

        if (m.tags.length === 0) {
            m.status = m.album >= 0 ? "album" : "ready";

            if (m.status === "ready") {
                UploadController.OnUploadEntryReady(m);
            } else {
                emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));
                UploadController.CheckQueue();
            }

            return;
        }

        m.busy = true;

        const tag = m.tags[0];
        const mediaId = m.mid;

        makeNamedApiRequest(REQUEST_PREFIX_UPLOAD + m.id, apiTagsTagMedia(mediaId, tag))
            .onSuccess((res) => {
                setLastUsedTag(res.id, "edit");

                m.tags.shift(); // Remove tag from list
                m.progress = m.tags.length;
                m.busy = false;

                emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                TagsController.AddTag(res.id, res.name);

                UploadController.AddNextTag(m);
            })
            .onCancel(() => {
                m.busy = false;
            })
            .onRequestError((err, handleErr) => {
                m.busy = false;
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);

                        m.error = "access-denied";
                        m.status = "error";
                        UploadController.OnEntryError(m);
                    },
                    invalidTagName: () => {
                        m.tags.shift(); // Invalid tag name
                        m.progress = m.tags.length;
                        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                        UploadController.AddNextTag(m);
                    },
                    badRequest: () => {
                        m.tags.shift(); // Invalid tag name
                        m.progress = m.tags.length;
                        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                        UploadController.AddNextTag(m);
                    },
                    accessDenied: () => {
                        m.error = "access-denied";
                        m.status = "error";
                        UploadController.OnEntryError(m);
                    },
                    temporalError: () => {
                        // Retry

                        m.retryTimer = setTimeout(() => {
                            UploadController.AddNextTag(m);
                        }, RETRY_DELAY_MS);
                    },
                });
            })
            .onUnexpectedError((err) => {
                m.busy = false;
                console.error(err);

                // Retry

                m.retryTimer = setTimeout(() => {
                    UploadController.AddNextTag(m);
                }, RETRY_DELAY_MS);
            });
    }

    /**
     * Inserts media into album
     * @param m The entry
     */
    private static InsertIntoAlbum(m: UploadEntry) {
        if (m.busy) {
            return;
        }

        m.retryTimer = null;

        m.busy = true;

        const albumId = m.album;
        const mediaId = m.mid;

        makeNamedApiRequest(REQUEST_PREFIX_UPLOAD + m.id, apiAlbumsAddMediaToAlbum(albumId, mediaId))
            .onSuccess(() => {
                m.busy = false;
                m.status = "ready";

                UploadController.OnUploadEntryReady(m);
            })
            .onCancel(() => {
                m.busy = false;
            })
            .onRequestError((err, handleErr) => {
                m.busy = false;
                handleErr(err, {
                    unauthorized: () => {
                        emitAppEvent(EVENT_NAME_UNAUTHORIZED);

                        m.error = "access-denied";
                        m.status = "error";
                        UploadController.OnEntryError(m);
                    },
                    badRequest: () => {
                        m.status = "ready";

                        UploadController.OnUploadEntryReady(m);
                    },
                    notFound: () => {
                        m.status = "ready";

                        UploadController.OnUploadEntryReady(m);
                    },
                    maxSizeReached: () => {
                        m.status = "ready";

                        UploadController.OnUploadEntryReady(m);
                    },
                    accessDenied: () => {
                        m.error = "access-denied";
                        m.status = "error";
                        UploadController.OnEntryError(m);
                    },
                    temporalError: () => {
                        // Retry

                        m.retryTimer = setTimeout(() => {
                            UploadController.InsertIntoAlbum(m);
                        }, RETRY_DELAY_MS);
                    },
                });
            })
            .onUnexpectedError((err) => {
                m.busy = false;
                console.error(err);

                // Retry

                m.retryTimer = setTimeout(() => {
                    UploadController.InsertIntoAlbum(m);
                }, RETRY_DELAY_MS);
            });
    }
}

/**
 * Minimizes upload entry
 * @param e The entry
 * @returns The minimized entry
 */
function minimizeEntry(e: UploadEntry): UploadEntryMin {
    return {
        id: e.id,
        name: e.name,
        size: e.size,
        mid: e.mid,
        status: e.status,
        error: e.error,
        progress: e.progress,
    };
}
