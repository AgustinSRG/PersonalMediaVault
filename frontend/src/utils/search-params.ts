// Search parameters utils

"use strict";

/**
 * Packs search params into an string
 * @param page The current page index
 * @param order The current order (asc or desc)
 * @returns The packed search params
 */
export function packSearchParams(page: number, order: string): string {
    if (page === 0 && order === "desc") {
        return "";
    }

    if (order === "desc") {
        return page + "";
    }

    return page + "-" + order;
}

/**
 * Search parameters
 */
export interface SearchParams {
    /**
     * Page number
     */
    page: number;

    /**
     * Order direction
     */
    order: "asc" | "desc";
}

/**
 * Unpacks the search parameters from a string
 * @param params The packed search params
 * @returns The search parameters (page index and order)
 */
export function unPackSearchParams(params: string): SearchParams {
    const res: SearchParams = {
        page: 0,
        order: "desc",
    };

    if (params) {
        const spl = params.split("-");
        res.page = parseInt(spl[0], 10) || 0;
        if (res.page < 0) {
            res.page = 0;
        }

        res.order = spl[1] !== "asc" ? "desc" : "asc";
    }

    return res;
}
