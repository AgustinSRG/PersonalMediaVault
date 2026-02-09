// Common request errors composable

"use strict";

import type { Ref } from "vue";
import { ref } from "vue";
import { useI18n } from "./use-i18n";
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/control/app-events";

/**
 * Composable that provides handlers for common errors
 */
export type CommonRequestErrorsComposable = {
    /**
     * Error message
     */
    error: Ref<string>;

    /**
     * Handler for 'unauthorized' error
     */
    unauthorized: () => void;

    /**
     * Handler for 'bad request' error
     */
    badRequest: () => void;

    /**
     * Handler for 'access denied' error
     */
    accessDenied: () => void;

    /**
     * Handler for internal server errors
     */
    serverError: () => void;

    /**
     * Handler for network errors
     */
    networkError: () => void;
};

/**
 * Gets an error message reference and handlers for common errors
 */
export function useCommonRequestErrors(): CommonRequestErrorsComposable {
    // Translation function
    const { $t } = useI18n();

    // Error message
    const error = ref("");

    // Handler for 'unauthorized' error
    const unauthorized = () => {
        error.value = $t("Access denied");
        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
    };

    // Handler for 'bad request' error
    const badRequest = () => {
        error.value = $t("Bad request");
    };

    // Handler for 'access denied' error
    const accessDenied = () => {
        error.value = $t("Access denied");
    };

    // Handler for internal server errors
    const serverError = () => {
        error.value = $t("Internal server error");
    };

    // Handler for network errors
    const networkError = () => {
        error.value = $t("Could not connect to the server");
    };

    return {
        error,
        unauthorized,
        badRequest,
        accessDenied,
        serverError,
        networkError,
    };
}
