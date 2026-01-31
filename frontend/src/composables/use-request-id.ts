// Composable to generate request IDs

"use strict";

import { clearNamedTimeout } from "@/utils/named-timeouts";
import { getUniqueStringId } from "@/utils/unique-id";
import { abortNamedApiRequest } from "@asanrom/request-browser";
import { onBeforeUnmount } from "vue";

/**
 * Generates an unique identifier for a request.
 * If the component unmounts, the named request or timeout
 * will be automatically cancelled.
 * @returns The generated unique request id
 */
export function useRequestId(): string {
    const requestId = getUniqueStringId();

    onBeforeUnmount(() => {
        clearNamedTimeout(requestId);
        abortNamedApiRequest(requestId);
    });

    return requestId;
}
