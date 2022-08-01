// Media API

import { GetAPIURL, RequestParams } from "@/utils/request";

export class MediaAPI {
    public static UploadMedia(title: string, file: File): RequestParams {
        const form = new FormData();
        form.append("file", file);

        return {
            method: "POST",
            url: GetAPIURL("/api/upload?title=" + encodeURIComponent(title)),
            form: form,
        };
    }

    public static GetMedia(id: number): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "")),
        };
    }

    public static ChangeMediaTitle(id: number, title: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/edit/title"),
            json: {
                title: title,
            },
        };
    }

    public static ChangeMediaDescription(id: number, description: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/edit/description"),
            json: {
                description: description,
            },
        };
    }

    public static ChangeExtraParams(id: number, forceStartBeginning: boolean): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/edit/extra"),
            json: {
                force_start_beginning: forceStartBeginning,
            },
        };
    }

    public static ChangeMediaThumbnail(id: number, thumbnail: File): RequestParams {
        const form = new FormData();
        form.append("file", thumbnail);
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/edit/thumbnail"),
            form: form,
        };
    }

    public static EncodeMedia(id: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/encode"),
        };
    }

    public static DeleteMedia(id: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/delete"),
        };
    }


    public static AddResolution(id: number, width: number, height: number, fps: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/resolution/add"),
            json: {
                width: width,
                height: height,
                fps: fps,
            },
        };
    }

    public static RemoveResolution(id: number, width: number, height: number, fps: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(id + "") + "/resolution/remove"),
            json: {
                width: width,
                height: height,
                fps: fps,
            },
        };
    }

    public static SetSubtitles(mediaId: number, id: string, name: string, srt: File): RequestParams {
        const form = new FormData();
        form.append("file", srt);
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(mediaId + "") + "/subtitles/set?id=" + encodeURIComponent(id) + "&name=" + encodeURIComponent(name)),
            form: form,
        };
    }

    public static RemoveSubtitles(mediaId: number, id: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/media/" + encodeURIComponent(mediaId + "") + "/subtitles/remove?id=" + encodeURIComponent(id)),
        };
    }
}
