// Edit media API

import { ImageNote } from "@/control/img-notes";
import { GetApiURL, RequestParams } from "@/utils/request";
import { MediaResolution, MediaSubtitle, MediaAudioTrack } from "./models";

export class EditMediaAPI {
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

    public static UploadAttachment(mediaId: number, attachment: File): RequestParams<MediaAudioTrack> {
        const form = new FormData();
        form.append("file", attachment);
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(mediaId + "") + "/attachments/add"),
            form: form,
        };
    }

    public static RenameAttachment(mediaId: number, id: number, name: string): RequestParams<MediaAudioTrack> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(mediaId + "") + "/attachments/rename"),
            json: {
                id: id,
                name: name,
            },
        };
    }

    public static RemoveAttachment(mediaId: number, id: number): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/media/" + encodeURIComponent(mediaId + "") + "/attachments/remove?id=" + encodeURIComponent(id + "")),
        };
    }
}