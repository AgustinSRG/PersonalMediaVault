// Media API

"use strict";

import { RequestParams } from "./request";
import { MediaData, MediaSizeStats } from "./models";
import { getApiURL } from "./utils";

export class MediaAPI {
    public static GetMedia(id: number): RequestParams<MediaData> {
        return {
            method: "GET",
            url: getApiURL("/api/media/" + encodeURIComponent(id + "")),
        };
    }

    public static GetMediaAlbums(id: number): RequestParams<number[]> {
        return {
            method: "GET",
            url: getApiURL("/api/media/" + encodeURIComponent(id + "") + "/albums"),
        };
    }

    public static GetMediaSizeStats(id: number): RequestParams<MediaSizeStats> {
        return {
            method: "GET",
            url: getApiURL("/api/media/" + encodeURIComponent(id + "") + "/size_stats"),
        };
    }
}
