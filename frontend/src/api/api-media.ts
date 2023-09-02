// Media API

import { GetApiURL, RequestParams } from "@/utils/request";
import { MediaData, MediaSizeStats } from "./models";

export class MediaAPI {
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
}
