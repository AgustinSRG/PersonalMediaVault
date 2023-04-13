// Tags API

import { GetApiURL, RequestParams } from "@/utils/request";

export class TagsAPI {
    public static GetTags(): RequestParams {
        return {
            method: "GET",
            url: GetApiURL("/api/tags"),
        };
    }

    public static TagMedia(media: number, tagName: string): RequestParams {
        return {
            method: "POST",
            url: GetApiURL("/api/tags/add"),
            json: {
                media_id: media,
                tag_name: tagName,
            },
        };
    }

    public static UntagMedia(media: number, tagId: number): RequestParams {
        return {
            method: "POST",
            url: GetApiURL("/api/tags/remove"),
            json: {
                media_id: media,
                tag_id: tagId,
            },
        };
    }
}
