// Upload global state

"use strict";

import { makeNamedApiRequest, abortNamedApiRequest } from "@asanrom/request-browser";
import { indicateTagCreation } from "./tags";
import { getMaxParallelUploads, setLastUsedTag, setMaxParallelUploads } from "../local-storage/app-preferences";
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
import { getCurrentAlbumMediaPositionContext, indicateAlbumMetadataChanged, loadAlbumNextPreFetch } from "./album";
import { getCurrentMediaId, loadCurrentMedia } from "./media";
import { addGlobalBusyCheck } from "./busy-state";
import { LOAD_RETRY_DELAY } from "@/constants";
import { removeExtensionFromFileName } from "@/utils/files";

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
 * Prefix for upload requests
 */
const REQUEST_PREFIX_UPLOAD = "upload-media-";

/**
 * Prefix for check requests
 */
const REQUEST_PREFIX_CHECK_ENCRYPTION = "check-media-encryption-";

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
 * Upload state
 */
const UploadState = {
    /**
     * Upload entries
     */
    entries: {
        pending: [] as UploadEntry[],
        ready: [] as UploadEntry[],
        error: [] as UploadEntry[],
    },

    /**
     * Map of upload entries
     * ID -> Entry
     */
    entriesMap: new Map<number, UploadEntry>(),

    /**
     * Upload queue
     */
    uploadQueue: [] as UploadEntry[],

    /**
     * Album queue
     * Album ID -> Entries to add to the album
     */
    albumQueue: new Map<number, UploadEntry[]>(),

    /**
     * Number of active uploads
     */
    uploadingCount: 0,

    /**
     * Max number of parallel active uploads
     */
    maxParallelUploads: getMaxParallelUploads(),

    /**
     * Counter to assign unique IDs to each entry
     */
    nextId: 0,
};

/**
 * Gets the list of pending entries
 * @returns The list of entries
 */
export function uploadGetPendingEntries(): UploadEntryMin[] {
    return UploadState.entries.pending.map(minimizeEntry);
}

/**
 * Gets the list of ready entries
 * @returns The list of entries
 */
export function uploadGetReadyEntries(): UploadEntryMin[] {
    return UploadState.entries.ready.map(minimizeEntry);
}

/**
 * Gets the list of ready entries
 * @returns The list of entries
 */
export function uploadGetErrorEntries(): UploadEntryMin[] {
    return UploadState.entries.error.map(minimizeEntry);
}

/**
 * Gets the max parallel uploads
 * @returns The max parallel uploads
 */
export function uploadGetMaxParallelUploads(): number {
    return UploadState.maxParallelUploads;
}

/**
 * Sets the max parallel uploads
 * @param maxParallelUploads The max parallel uploads
 */
export function uploadSetMaxParallelUploads(maxParallelUploads: number) {
    UploadState.maxParallelUploads = maxParallelUploads;
    setMaxParallelUploads(maxParallelUploads);
    checkQueue();
}

/**
 * Checks the upload queue
 */
function checkQueue() {
    while (UploadState.uploadQueue.length > 0 && UploadState.uploadingCount < UploadState.maxParallelUploads) {
        uploadMedia(UploadState.uploadQueue.shift());
    }

    UploadState.albumQueue.forEach((queue) => {
        if (queue.length === 0) {
            return;
        }

        const headEntry = queue[0];

        if (headEntry.busy || headEntry.status !== "album") {
            return;
        }

        insertIntoAlbum(headEntry);
    });
}

/**
 * Adds entry to album queue
 * @param entry The entry
 */
function addToAlbumQueue(entry: UploadEntry) {
    if (entry.album < 0) {
        return;
    }

    if (UploadState.albumQueue.has(entry.album)) {
        const q = UploadState.albumQueue.get(entry.album);
        q.push(entry);
    } else {
        UploadState.albumQueue.set(entry.album, [entry]);
    }
}

/**
 * Removes entry from album queue
 * @param entry The upload entry
 */
function removeFromAlbumQueue(entry: UploadEntry) {
    if (entry.album < 0) {
        return;
    }

    if (UploadState.albumQueue.has(entry.album)) {
        const q = UploadState.albumQueue.get(entry.album);
        removeFromArray(q, entry);

        if (q.length === 0) {
            UploadState.albumQueue.delete(entry.album);
        }
    }
}

/**
 * Adds a file to the upload queue, creating a new entry
 * @param file The file to upload
 * @param album The album to add the media into. Set to -1 for no album.
 * @param tags The list of tags to add to the media.
 * @returns The created entry ID
 */
export function uploadMediaFile(file: File, album: number, tags: string[]): number {
    UploadState.nextId++;
    const id = UploadState.nextId;

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

    UploadState.entriesMap.set(entry.id, entry);
    UploadState.entries.pending.push(entry);
    UploadState.uploadQueue.push(entry);

    if (album >= 0) {
        if (UploadState.albumQueue.has(album)) {
            const q = UploadState.albumQueue.get(album);
            q.push(entry);
        } else {
            UploadState.albumQueue.set(album, [entry]);
        }
    }

    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_NEW, minimizeEntry(entry));

    checkQueue();

    return id;
}

/**
 * Retries uploading a media file, when it resulted in error
 * @param id The upload entry ID
 */
export function retryMediaUpload(id: number) {
    const entry = UploadState.entriesMap.get(id);

    if (!entry || entry.status !== "error") {
        return;
    }

    entry.error = "";
    entry.status = "pending";
    entry.progress = 0;

    removeFromArray(UploadState.entries.error, entry);
    UploadState.entries.pending.push(entry);

    UploadState.uploadQueue.push(entry);

    addToAlbumQueue(entry);

    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_RETRY, minimizeEntry(entry));
}

/**
 * Removes an entry, aborting any requests.
 * @param id The entry ID
 */
export function removeUploadEntry(id: number) {
    const entry = UploadState.entriesMap.get(id);

    if (!entry) {
        return;
    }

    UploadState.entriesMap.delete(entry.id);

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
            removeFromArray(UploadState.entries.ready, entry);
            break;
        case "error":
            removeFromArray(UploadState.entries.error, entry);
            break;
        default:
            removeFromArray(UploadState.entries.pending, entry);
            removeFromArray(UploadState.uploadQueue, entry);
    }

    removeFromAlbumQueue(entry);

    // Decrease the uploading count in case if was in the encrypting phase
    // In the uploading phase, since it is a single request, it is handled in the cancel callback

    if (entry.status === "encrypting") {
        UploadState.uploadingCount--;
    }

    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED, minimizeEntry(entry));
}

/**
 * Clears any completed uploads from the list
 */
export function clearUploadList() {
    const completedUploads = UploadState.entries.ready.concat(UploadState.entries.error);

    for (const entry of completedUploads) {
        UploadState.entriesMap.delete(entry.id);
    }

    UploadState.entries.ready = [];
    UploadState.entries.error = [];

    emitAppEvent(EVENT_NAME_UPLOAD_LIST_CLEAR);
}

/**
 * Cancels all pending uploads
 */
export function cancelAllPendingUploads() {
    const pendingUploads = UploadState.entries.pending;

    for (const entry of pendingUploads) {
        removeUploadEntry(entry.id);
    }

    const completedUploads = UploadState.entries.ready.concat(UploadState.entries.error);

    for (const entry of completedUploads) {
        UploadState.entriesMap.delete(entry.id);
    }

    UploadState.entries.pending = [];
    UploadState.entries.ready = [];
    UploadState.entries.error = [];

    emitAppEvent(EVENT_NAME_UPLOAD_LIST_CLEAR);
}

/**
 * Checks a completed list for its max length
 * @param list The completed list (ready or error)
 */
function checkCompletedList(list: UploadEntry[]) {
    if (list.length > MAX_COMPLETED_UPLOADS) {
        const toRemove = list.shift();

        UploadState.entriesMap.delete(toRemove.id);

        emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_REMOVED, minimizeEntry(toRemove));
    }
}

/**
 * Call when entry is ready
 * @param m The upload entry
 */
function onUploadEntryReady(m: UploadEntry) {
    removeFromArray(UploadState.entries.pending, m);
    UploadState.entries.ready.push(m);

    removeFromAlbumQueue(m);

    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_READY, minimizeEntry(m));

    checkQueue();

    if (getCurrentMediaId() === m.mid) {
        loadCurrentMedia();
    }

    const albumMediaPositionContext = getCurrentAlbumMediaPositionContext();

    if (albumMediaPositionContext.next && albumMediaPositionContext.next.id === m.mid) {
        loadAlbumNextPreFetch();
    }

    if (m.album !== -1) {
        indicateAlbumMetadataChanged(m.album, true);
    }

    checkCompletedList(UploadState.entries.ready);
}

/**
 * Call when an entry reaches the error status
 * Moves the entry and emits the event
 * @param m The entry
 */
function onEntryError(m: UploadEntry) {
    removeFromArray(UploadState.entries.pending, m);
    UploadState.entries.error.push(m);

    removeFromAlbumQueue(m);

    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_ERROR, minimizeEntry(m));

    checkQueue();

    checkCompletedList(UploadState.entries.error);
}

/**
 * Uploads the media file
 * @param m The entry
 */
function uploadMedia(m: UploadEntry) {
    UploadState.uploadingCount++;

    m.status = "uploading";
    m.progress = 0;
    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

    makeNamedApiRequest(REQUEST_PREFIX_UPLOAD + m.id, apiUploadMedia(removeExtensionFromFileName(m.name), m.file, -1))
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
                checkEncryptionStatus(m);
            }, ENCRYPTION_CHECK_INTERVAL_MS);
        })
        .onCancel(() => {
            UploadState.uploadingCount--;
            removeUploadEntry(m.id);
            checkQueue();
        })
        .onRequestError((err, handleErr) => {
            UploadState.uploadingCount--;

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

            onEntryError(m);
        })
        .onUnexpectedError((err) => {
            UploadState.uploadingCount--;

            m.error = "no-internet";
            console.error(err);
            m.status = "error";

            onEntryError(m);
        });
}

/**
 * Checks encryption status of the media
 * @param m The entry
 */
function checkEncryptionStatus(m: UploadEntry) {
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

                UploadState.uploadingCount--;

                if (m.status === "ready") {
                    onUploadEntryReady(m);
                } else {
                    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));
                    checkQueue();

                    if (m.status === "tag") {
                        addNextTag(m);
                    }
                }
            } else {
                m.progress = Math.max(m.progress, media.ready_p);
                emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                m.retryTimer = setTimeout(() => {
                    checkEncryptionStatus(m);
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

                    UploadState.uploadingCount--;

                    onEntryError(m);
                },
                temporalError: () => {
                    // Retry

                    m.retryTimer = setTimeout(() => {
                        checkEncryptionStatus(m);
                    }, LOAD_RETRY_DELAY);
                },
            });
        })
        .onUnexpectedError((err) => {
            m.busy = false;
            console.error(err);

            // Retry

            m.retryTimer = setTimeout(() => {
                checkEncryptionStatus(m);
            }, LOAD_RETRY_DELAY);
        });
}

/**
 * Adds next tag to the media
 * @param m The entry
 */
function addNextTag(m: UploadEntry) {
    if (m.busy) {
        return;
    }

    m.retryTimer = null;

    if (m.tags.length === 0) {
        m.status = m.album >= 0 ? "album" : "ready";

        if (m.status === "ready") {
            onUploadEntryReady(m);
        } else {
            emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));
            checkQueue();
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

            indicateTagCreation(res.id, res.name);

            addNextTag(m);
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
                    onEntryError(m);
                },
                invalidTagName: () => {
                    m.tags.shift(); // Invalid tag name
                    m.progress = m.tags.length;
                    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                    addNextTag(m);
                },
                badRequest: () => {
                    m.tags.shift(); // Invalid tag name
                    m.progress = m.tags.length;
                    emitAppEvent(EVENT_NAME_UPLOAD_LIST_ENTRY_PROGRESS, minimizeEntry(m));

                    addNextTag(m);
                },
                accessDenied: () => {
                    m.error = "access-denied";
                    m.status = "error";
                    onEntryError(m);
                },
                temporalError: () => {
                    // Retry

                    m.retryTimer = setTimeout(() => {
                        addNextTag(m);
                    }, LOAD_RETRY_DELAY);
                },
            });
        })
        .onUnexpectedError((err) => {
            m.busy = false;
            console.error(err);

            // Retry

            m.retryTimer = setTimeout(() => {
                addNextTag(m);
            }, LOAD_RETRY_DELAY);
        });
}

/**
 * Inserts media into album
 * @param m The entry
 */
function insertIntoAlbum(m: UploadEntry) {
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

            onUploadEntryReady(m);
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
                    onEntryError(m);
                },
                badRequest: () => {
                    m.status = "ready";

                    onUploadEntryReady(m);
                },
                notFound: () => {
                    m.status = "ready";

                    onUploadEntryReady(m);
                },
                maxSizeReached: () => {
                    m.status = "ready";

                    onUploadEntryReady(m);
                },
                accessDenied: () => {
                    m.error = "access-denied";
                    m.status = "error";

                    onEntryError(m);
                },
                temporalError: () => {
                    // Retry

                    m.retryTimer = setTimeout(() => {
                        insertIntoAlbum(m);
                    }, LOAD_RETRY_DELAY);
                },
            });
        })
        .onUnexpectedError((err) => {
            m.busy = false;
            console.error(err);

            // Retry

            m.retryTimer = setTimeout(() => {
                insertIntoAlbum(m);
            }, LOAD_RETRY_DELAY);
        });
}

// Global busy check for pending uploads
addGlobalBusyCheck(() => uploadGetPendingEntries().length > 0);
