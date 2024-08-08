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
    order: "asc" | "desc" | "rand";
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

        switch (spl[1]) {
            case "asc":
            case "rand":
                res.order = spl[1];
                break;
            default:
                res.order = "desc";
        }
    }

    return res;
}

/**
 * Simplifies the order
 * @param order Order parameter
 * @returns Simplified order parameter
 */
export function orderSimple(order: "asc" | "desc" | "rand"): "asc" | "desc" {
    return order === "asc" ? "asc" : "desc";
}
