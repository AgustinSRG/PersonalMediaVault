// Search API

import { GenerateURIQuery, GetApiURL, RequestParams } from "@/utils/request";
import { SearchResults, RandomResults } from "./models";

export class SearchAPI {
    public static Search(tag: string, order: string, page: number, pageSize: number): RequestParams<SearchResults> {
        return {
            method: "GET",
            url: GetApiURL(
                "/api/search" +
                    GenerateURIQuery({
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
            url: GetApiURL(
                "/api/random" +
                    GenerateURIQuery({
                        tag: tag,
                        seed: seed + "",
                        page_size: pageSize + "",
                    }),
            ),
        };
    }
}
