// Albums API

import { GetAPIURL, RequestParams } from "@/utils/request";

export class AmbumsAPI {
    public static GetAlbums(): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/albums"),
        };
    }

    public static GetAlbum(id: number): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/albums/" + encodeURIComponent(id + "")),
        };
    }

    public static CreateAlbum(name: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/albums/"),
            json: {
                name: name,
            },
        };
    }

    public static DeleteAlbum(id: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/albums/" + encodeURIComponent(id + "") + "/delete"),
        };
    }

    public static RenameAlbum(id: number, name: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/albums/" + encodeURIComponent(id + "") + "/delete"),
            json: {
                name: name,
            },
        };
    }

    public static SetAlbumOrder(id: number, list: number[]): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/albums/" + encodeURIComponent(id + "") + "/set"),
            json: {
                list: list,
            },
        };
    }

    public static AddMediaToAlbum(id: number, media: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/albums/" + encodeURIComponent(id + "") + "/add"),
            json: {
                media_id: media,
            },
        };
    }

    public static RemoveMediaFromAlbum(id: number, media: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/albums/" + encodeURIComponent(id + "") + "/remove"),
            json: {
                media_id: media,
            },
        };
    }
}
