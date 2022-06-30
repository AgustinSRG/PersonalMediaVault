// Tags API

import { GetAPIURL, RequestParams } from "@/utils/request";

export class TagsAPI {
    public static GetTags(): RequestParams {
        return {
            method: "GET",
            url: GetAPIURL("/api/tags"),
        };
    }

    public static TagMedia(media: number, tagName: string): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/tags/add"),
            json: {
                media_id: media,
                tag_name: tagName,
            },
        };
    }

    public static UntagMedia(media: number, tagId: number): RequestParams {
        return {
            method: "POST",
            url: GetAPIURL("/api/tags/remove"),
            json: {
                media_id: media,
                tag_id: tagId,
            },
        };
    }
}
