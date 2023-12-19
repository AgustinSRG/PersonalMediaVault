// Edit media API

"use strict";

import { CommonAuthenticatedErrorHandler, RequestErrorHandler, RequestParams } from "@asanrom/request-browser";
import { MediaResolution, MediaSubtitle, MediaAudioTrack, MediaTimeSlice, MediaAttachment } from "./models";
import { ImageNote } from "@/utils/notes-format";
import { API_PREFIX, getApiURL } from "@/utils/api";

const API_GROUP_PREFIX = "/media";

/**
 * Error handler for all media edit APIs
 */
export type MediaEditApiErrorHandler = CommonAuthenticatedErrorHandler & {
    /**
     * Error: Access denied
     */
    accessDenied: () => void;

    /**
     * Error: Not found
     */
    notFound: () => void;
};

/**
 * Error handler for edit title API
 */
export type ChangeTitleErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Invalid title
     */
    invalidTitle: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Changes media title
 * @param id Media ID
 * @param title New title
 * @returns The request parameters
 */
export function apiMediaChangeMediaTitle(id: number, title: string): RequestParams<void, ChangeTitleErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/edit/title`),
        json: {
            title: title,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_TITLE", handler.invalidTitle)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for edit description API
 */
export type ChangeDescriptionErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Invalid description
     */
    invalidDescription: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Changes media description
 * @param id Media ID
 * @param description New description
 * @returns The request parameters
 */
export function apiMediaChangeMediaDescription(id: number, description: string): RequestParams<void, ChangeDescriptionErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/edit/description`),
        json: {
            description: description,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_DESCRIPTION", handler.invalidDescription)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for edit extra parameters API
 */
export type ChangeExtraParamsErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Changes media extra parameters
 * @param id Media ID
 * @param forceStartBeginning True to force start from the beginning on load
 * @returns The request parameters
 */
export function apiMediaChangeExtraParams(id: number, forceStartBeginning: boolean): RequestParams<void, ChangeExtraParamsErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/edit/extra`),
        json: {
            force_start_beginning: forceStartBeginning,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for edit time slices API
 */
export type ChangeTimeSlicesErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Changes time slices
 * @param id Media ID
 * @param time_slices List of time slices
 * @returns The request parameters
 */
export function apiMediaChangeTimeSlices(id: number, time_slices: MediaTimeSlice[]): RequestParams<void, ChangeTimeSlicesErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/edit/time_slices`),
        json: time_slices,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for edit image notes API
 */
export type SetImageNotesErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Sets image notes
 * @param id Media ID
 * @param notes List of image notes
 * @returns The request parameters
 */
export function apiMediaSetNotes(id: number, notes: ImageNote[]): RequestParams<void, SetImageNotesErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/edit/notes`),
        json: notes,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for set extended description API
 */
export type SetExtendedDescriptionErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Sets extended description
 * @param id Media ID
 * @param extendedDesc Extended description
 * @returns The request parameters
 */
export function apiMediaSetExtendedDescription(id: number, extendedDesc: string): RequestParams<void, SetExtendedDescriptionErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/edit/ext_desc`),
        json: {
            ext_desc: extendedDesc,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for set extended description API
 */
export type SetThumbnailErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Invalid thumbnail
     */
    invalidThumbnail: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Response of the thumbnail change API
 */
export interface ChangeThumbnailResponse {
    /**
     * New thumbnail URL
     */
    url: string;
}

/**
 * Sets thumbnail
 * @param id Media ID
 * @param thumbnail Thumbnail file to upload
 * @returns The request parameters
 */
export function apiMediaChangeMediaThumbnail(
    id: number,
    thumbnail: File,
): RequestParams<ChangeThumbnailResponse, SetThumbnailErrorHandler> {
    const form = new FormData();
    form.append("file", thumbnail);
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/edit/thumbnail`),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_THUMBNAIL", handler.invalidThumbnail)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Re-encodes media
 * @param id Media ID
 * @returns The request parameters
 */
export function apiMediaEncodeMedia(id: number): RequestParams<void, MediaEditApiErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/encode`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for set extended description API
 */
export type ReplaceMediaErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Invalid media file
     */
    invalidMedia: () => void;

    /**
     * Error: invalid media type
     */
    invalidMediaType: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Replaces media
 * @param id Media ID
 * @param file Media file to upload
 * @returns The request parameters
 */
export function apiMediaReplaceMedia(
    id: number,
    file: File,
): RequestParams<void, ReplaceMediaErrorHandler> {
    const form = new FormData();
    form.append("file", file);
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/replace`),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_MEDIA", handler.invalidMedia)
                .add(400, "INVALID_MEDIA_TYPE", handler.invalidMediaType)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Deletes media
 * @param id Media ID
 * @returns The request parameters
 */
export function apiMediaDeleteMedia(id: number): RequestParams<void, MediaEditApiErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/delete`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for add resolution API
 */
export type AddResolutionErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Duplicated resolution
     */
    duplicatedResolution: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Adds resolution
 * @param id Media ID
 * @param width Width
 * @param height Height
 * @param fps Frames per second
 * @returns The request parameters
 */
export function apiMediaAddResolution(
    id: number,
    width: number,
    height: number,
    fps: number,
): RequestParams<MediaResolution, AddResolutionErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/resolution/add`),
        json: {
            width: width,
            height: height,
            fps: fps,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "DUPLICATED_RESOLUTION", handler.duplicatedResolution)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for remove resolution API
 */
export type RemoveResolutionErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Removes resolution
 * @param id Media ID
 * @param width Width
 * @param height Height
 * @param fps Frames per second
 * @returns The request parameters
 */
export function apiMediaRemoveResolution(
    id: number,
    width: number,
    height: number,
    fps: number,
): RequestParams<void, RemoveResolutionErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(id + "")}/resolution/remove`),
        json: {
            width: width,
            height: height,
            fps: fps,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for set subtitles API
 */
export type SetSubtitlesErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Invalid SRT file
     */
    invalidSRT: () => void;

    /**
     * Error: Invalid ID
     */
    invalidId: () => void;

    /**
     * Error: Invalid name
     */
    invalidName: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;

    /**
     * Error: File too large (max 10MB)
     */
    fileTooLarge: () => void;
};

/**
 * Sets subtitles
 * @param mediaId Media ID
 * @param id Subtitles ID
 * @param name Subtitles name
 * @param srt SubRip file
 * @returns The request parameters
 */
export function apiMediaSetSubtitles(
    mediaId: number,
    id: string,
    name: string,
    srt: File,
): RequestParams<MediaSubtitle, SetSubtitlesErrorHandler> {
    const form = new FormData();
    form.append("file", srt);
    return {
        method: "POST",
        url: getApiURL(
            `${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(mediaId + "")}/subtitles/set?id=${encodeURIComponent(
                id,
            )}&name=${encodeURIComponent(name)}`,
        ),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_SRT", handler.invalidSRT)
                .add(400, "INVALID_ID", handler.invalidId)
                .add(400, "INVALID_NAME", handler.invalidName)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(413, "*", handler.fileTooLarge)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for remove subtitles API
 */
export type RemoveSubtitlesErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Removes subtitles file
 * @param mediaId Media ID
 * @param id Subtitles file ID
 * @returns The request parameters
 */
export function apiMediaRemoveSubtitles(mediaId: number, id: string): RequestParams<void, RemoveSubtitlesErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(
            `${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(mediaId + "")}/subtitles/remove?id=${encodeURIComponent(id)}`,
        ),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for set audio track API
 */
export type SetAudioTrackErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Invalid audio file
     */
    invalidAudio: () => void;

    /**
     * Error: Invalid ID
     */
    invalidId: () => void;

    /**
     * Error: Invalid name
     */
    invalidName: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Sets extra audio track
 * @param mediaId Media ID
 * @param id Audio track ID
 * @param name Audio track name
 * @param audio Audio file
 * @returns The request parameters
 */
export function apiMediaSetAudioTrack(
    mediaId: number,
    id: string,
    name: string,
    audio: File,
): RequestParams<MediaAudioTrack, SetAudioTrackErrorHandler> {
    const form = new FormData();
    form.append("file", audio);
    return {
        method: "POST",
        url: getApiURL(
            `${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(mediaId + "")}/audios/set?id=${encodeURIComponent(
                id,
            )}&name=${encodeURIComponent(name)}`,
        ),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_AUDIO", handler.invalidAudio)
                .add(400, "INVALID_ID", handler.invalidId)
                .add(400, "INVALID_NAME", handler.invalidName)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for remove audio track API
 */
export type RemoveAudioTrackErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Removes extra audio track
 * @param mediaId Media ID
 * @param id Audio track ID
 * @returns The request parameters
 */
export function apiMediaRemoveAudioTrack(mediaId: number, id: string): RequestParams<void, RemoveAudioTrackErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(mediaId + "")}/audios/remove?id=${encodeURIComponent(id)}`),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for upload attachment API
 */
export type UploadAttachmentErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Uploads attachment
 * @param mediaId Media ID
 * @param attachment Attachment file
 * @returns The request parameters
 */
export function apiMediaUploadAttachment(mediaId: number, attachment: File): RequestParams<MediaAttachment, UploadAttachmentErrorHandler> {
    const form = new FormData();
    form.append("file", attachment);
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(mediaId + "")}/attachments/add`),
        form: form,
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for rename attachment API
 */
export type RenameAttachmentErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Invalid name
     */
    invalidName: () => void;

    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Renames attachment
 * @param mediaId Media ID
 * @param id Attachment ID
 * @param name Attachment name
 * @returns The request parameters
 */
export function apiMediaRenameAttachment(
    mediaId: number,
    id: number,
    name: string,
): RequestParams<MediaAttachment, RenameAttachmentErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(`${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(mediaId + "")}/attachments/rename`),
        json: {
            id: id,
            name: name,
        },
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "INVALID_NAME", handler.invalidName)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}

/**
 * Error handler for remove attachment API
 */
export type RemoveAttachmentErrorHandler = MediaEditApiErrorHandler & {
    /**
     * Error: Bad request
     */
    badRequest: () => void;
};

/**
 * Removes attachment
 * @param mediaId Media ID
 * @param id Attachment ID
 * @returns The request parameters
 */
export function apiMediaRemoveAttachment(mediaId: number, id: number): RequestParams<void, RemoveAttachmentErrorHandler> {
    return {
        method: "POST",
        url: getApiURL(
            `${API_PREFIX}${API_GROUP_PREFIX}/${encodeURIComponent(mediaId + "")}/attachments/remove?id=${encodeURIComponent(id + "")}`,
        ),
        handleError: (err, handler) => {
            new RequestErrorHandler()
                .add(401, "*", handler.unauthorized)
                .add(400, "*", handler.badRequest)
                .add(403, "*", handler.accessDenied)
                .add(404, "*", handler.notFound)
                .add(500, "*", "serverError" in handler ? handler.serverError : handler.temporalError)
                .add("*", "*", "networkError" in handler ? handler.networkError : handler.temporalError)
                .handle(err);
        },
    };
}
