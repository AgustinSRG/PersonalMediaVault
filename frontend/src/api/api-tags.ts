// Tags API

"use strict";

import { RequestParams } from "./request";
import { MediaTag } from "./models";
import { getApiURL } from "./utils";

export class TagsAPI {
    public static GetTags(): RequestParams<MediaTag[]> {
        return {
            method: "GET",
            url: getApiURL("/api/tags"),
        };
    }

    public static TagMedia(media: number, tagName: string): RequestParams<MediaTag> {
        return {
            method: "POST",
            url: getApiURL("/api/tags/add"),
            json: {
                media_id: media,
                tag_name: tagName,
            },
        };
    }

    public static UntagMedia(media: number, tagId: number): RequestParams<{ removed: boolean }> {
        return {
            method: "POST",
            url: getApiURL("/api/tags/remove"),
            json: {
                media_id: media,
                tag_id: tagId,
            },
        };
    }
}
