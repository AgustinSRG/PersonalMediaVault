// Search API

import { GenerateURIQuery, GetApiURL, RequestParams } from "@/utils/request";

export class SearchAPI {
    public static Search(tag: string, order: string, page: number, pageSize: number): RequestParams {
        return {
            method: "GET",
            url: GetApiURL("/api/search" + GenerateURIQuery({
            
                tag: tag,
                order: order,
                page_index: page + "",
                page_size: pageSize + "",
            })),
        };
    }

    public static Random(tag: string, seed: number, pageSize: number): RequestParams {
        return {
            method: "GET",
            url: GetApiURL("/api/random" + GenerateURIQuery({
            
                tag: tag,
                seed: seed + "",
                page_size: pageSize + "",
            })),
        };
    }
}
