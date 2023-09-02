// Albums API

import { GetApiURL, RequestParams } from "@/utils/request";
import { AlbumListItem, AlbumListItemMin, Album } from "./models";

export class AlbumsAPI {
    public static GetAlbums(): RequestParams<AlbumListItem[]> {
        return {
            method: "GET",
            url: GetApiURL("/api/albums"),
        };
    }

    public static GetAlbumsMin(): RequestParams<AlbumListItemMin[]> {
        return {
            method: "GET",
            url: GetApiURL("/api/albums?mode=min"),
        };
    }

    public static GetAlbum(id: number): RequestParams<Album> {
        return {
            method: "GET",
            url: GetApiURL("/api/albums/" + encodeURIComponent(id + "")),
        };
    }

    public static CreateAlbum(name: string): RequestParams<{ album_id: number }> {
        return {
            method: "POST",
            url: GetApiURL("/api/albums"),
            json: {
                name: name,
            },
        };
    }

    public static DeleteAlbum(id: number): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/albums/" + encodeURIComponent(id + "") + "/delete"),
        };
    }

    public static RenameAlbum(id: number, name: string): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/albums/" + encodeURIComponent(id + "") + "/rename"),
            json: {
                name: name,
            },
        };
    }

    public static SetAlbumOrder(id: number, list: number[]): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/albums/" + encodeURIComponent(id + "") + "/set"),
            json: {
                list: list,
            },
        };
    }

    public static AddMediaToAlbum(id: number, media: number): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/albums/" + encodeURIComponent(id + "") + "/add"),
            json: {
                media_id: media,
            },
        };
    }

    public static RemoveMediaFromAlbum(id: number, media: number): RequestParams<void> {
        return {
            method: "POST",
            url: GetApiURL("/api/albums/" + encodeURIComponent(id + "") + "/remove"),
            json: {
                media_id: media,
            },
        };
    }
}
