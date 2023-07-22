// Media API

import { ImageNote } from "@/control/img-notes";
import { GetApiURL, RequestParams } from "@/utils/request";

export interface MediaListItem {
    id: number;
    type: 0 | 1 | 2 | 3;
    title: string;
    description: string;
    tags: number[];
    thumbnail: string;
    duration: number;
}

export interface MediaData {
    id: number;
    type: 0 | 1 | 2 | 3;
    title: string;
    description: string;
    tags: number[];
    upload_time: number;
    thumbnail: string;
    duration: number;
    width: number;
    height: number;
    fps: number;
    ready: boolean;
    ready_p: number;
    encoded: boolean;
    task: number;
    url: string;
    video_previews: string;
    video_previews_interval: number;
    resolutions: MediaResolution[];
    subtitles: MediaSubtitle[];
    audios: MediaAudioTrack[];
    force_start_beginning: boolean;
    img_notes: boolean;
    img_notes_url: string;
    ext_desc_url: string;
    time_slices: {
        time: number;
        name: string;
    }[];
}

export interface MediaResolution {
    width: number;
    height: number;
    fps: number;
    ready: boolean;
    task: number;
    url: string;
}

export interface MediaSubtitle {
    id: string;
    name: string;
    url: string;
}

export interface MediaAudioTrack {
    id: string;
    name: string;
    url: string;
}

export interface MediaSizeStats {
    meta_size: number;
    assets: {
        id: number;
        type: "s" | "m";
        name: string;
        size: number;
    }[];
}
export class MediaAPI {
    public static UploadMedia(title: string, file: File, album: number): RequestParams<{ media_id: number }> {
        const form = new FormData();
        form.append("file", file);

        return {
            method: "POST",
            url: GetApiURL("/api/upload?title=" + encodeURIComponent(title) + "&album=" + encodeURIComponent(album + "")),
            form: form,
        };
    }

    public static GetMedia(id: number): RequestParams<MediaData> {
        return {
            method: "GET",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "")),
        };
    }

    public static GetMediaAlbums(id: number): RequestParams<number[]> {
        return {
            method: "GET",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/albums"),
        };
    }

    public static GetMediaSizeStats(id: number): RequestParams<MediaSizeStats> {
        return {
            method: "GET",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/size_stats"),
        };
    }

    public static ChangeMediaTitle(id: number, title: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/edit/title"),
            json: {
                title: title,
            },
        };
    }

    public static ChangeMediaDescription(id: number, description: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/edit/description"),
            json: {
                description: description,
            },
        };
    }

    public static ChangeExtraParams(id: number, forceStartBeginning: boolean): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/edit/extra"),
            json: {
                force_start_beginning: forceStartBeginning,
            },
        };
    }

    public static ChangeTimeSlices(id: number, time_slices: { time: number; name: string }[]): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/edit/time_slices"),
            json: time_slices,
        };
    }

    public static SetNotes(id: number, notes: ImageNote[]): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/edit/notes"),
            json: notes,
        };
    }

    public static SetExtendedDescription(id: number, extendedDesc: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/edit/ext_desc"),
            json: {
                ext_desc: extendedDesc,
            },
        };
    }

    public static ChangeMediaThumbnail(id: number, thumbnail: File): RequestParams<{ url: string }> {
        const form = new FormData();
        form.append("file", thumbnail);
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/edit/thumbnail"),
            form: form,
        };
    }

    public static EncodeMedia(id: number): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/encode"),
        };
    }

    public static DeleteMedia(id: number): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/delete"),
        };
    }

    public static AddResolution(id: number, width: number, height: number, fps: number): RequestParams<MediaResolution> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/resolution/add"),
            json: {
                width: width,
                height: height,
                fps: fps,
            },
        };
    }

    public static RemoveResolution(id: number, width: number, height: number, fps: number): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(id + "") + "/resolution/remove"),
            json: {
                width: width,
                height: height,
                fps: fps,
            },
        };
    }

    public static SetSubtitles(mediaId: number, id: string, name: string, srt: File): RequestParams<MediaSubtitle> {
        const form = new FormData();
        form.append("file", srt);
        return {
            method: "POST",
            url: GetApiURL(
                "/api/media/" +
                    encodeURIComponent(mediaId + "") +
                    "/subtitles/set?id=" +
                    encodeURIComponent(id) +
                    "&name=" +
                    encodeURIComponent(name),
            ),
            form: form,
        };
    }

    public static RemoveSubtitles(mediaId: number, id: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(mediaId + "") + "/subtitles/remove?id=" + encodeURIComponent(id)),
        };
    }

    public static SetAudioTrack(mediaId: number, id: string, name: string, audio: File): RequestParams<MediaAudioTrack> {
        const form = new FormData();
        form.append("file", audio);
        return {
            method: "POST",
            url: GetApiURL(
                "/api/media/" +
                    encodeURIComponent(mediaId + "") +
                    "/audios/set?id=" +
                    encodeURIComponent(id) +
                    "&name=" +
                    encodeURIComponent(name),
            ),
            form: form,
        };
    }

    public static RemoveAudioTrack(mediaId: number, id: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(mediaId + "") + "/audios/remove?id=" + encodeURIComponent(id)),
        };
    }
}
