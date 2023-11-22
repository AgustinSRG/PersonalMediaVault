// Search API

"use strict";

import { RequestParams } from "@asanrom/request-browser";
import { SearchResults, RandomResults } from "./models";
import { generateURIQuery, getApiURL } from "@/utils/api";

export class SearchAPI {
    public static Search(tag: string, order: string, page: number, pageSize: number): RequestParams<SearchResults> {
        return {
            method: "GET",
            url: getApiURL(
                "/api/search" +
                    generateURIQuery({
                        tag: tag,
                        order: order,
                        page_index: page + "",
                        page_size: pageSize + "",
                    }),
            ),
        };
    }

    public static Random(tag: string, seed: number, pageSize: number): RequestParams<RandomResults> {
        return {
            method: "GET",
            url: getApiURL(
                "/api/random" +
                    generateURIQuery({
                        tag: tag,
                        seed: seed + "",
                        page_size: pageSize + "",
                    }),
            ),
        };
    }
}
