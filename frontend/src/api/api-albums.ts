// Albums API

"use strict";

import { RequestParams } from "@asanrom/request-browser";
import { AlbumListItem, AlbumListItemMin, Album } from "./models";
import { getApiURL } from "@/utils/api";

export class AlbumsAPI {
    public static GetAlbums(): RequestParams<AlbumListItem[]> {
        return {
            method: "GET",
            url: getApiURL("/api/albums"),
        };
    }

    public static GetAlbumsMin(): RequestParams<AlbumListItemMin[]> {
        return {
            method: "GET",
            url: getApiURL("/api/albums?mode=min"),
        };
    }

    public static GetAlbum(id: number): RequestParams<Album> {
        return {
            method: "GET",
            url: getApiURL("/api/albums/" + encodeURIComponent(id + "")),
        };
    }

    public static CreateAlbum(name: string): RequestParams<{ album_id: number }> {
        return {
            method: "POST",
            url: getApiURL("/api/albums"),
            json: {
                name: name,
            },
        };
    }

    public static DeleteAlbum(id: number): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/albums/" + encodeURIComponent(id + "") + "/delete"),
        };
    }

    public static RenameAlbum(id: number, name: string): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/albums/" + encodeURIComponent(id + "") + "/rename"),
            json: {
                name: name,
            },
        };
    }

    public static AddMediaToAlbum(id: number, media: number): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/albums/" + encodeURIComponent(id + "") + "/add"),
            json: {
                media_id: media,
            },
        };
    }

    public static RemoveMediaFromAlbum(id: number, media: number): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/albums/" + encodeURIComponent(id + "") + "/remove"),
            json: {
                media_id: media,
            },
        };
    }

    public static MoveMediaInAlbum(id: number, media: number, position: number): RequestParams<void> {
        return {
            method: "POST",
            url: getApiURL("/api/albums/" + encodeURIComponent(id + "") + "/move"),
            json: {
                media_id: media,
                position: position,
            },
        };
    }
}
