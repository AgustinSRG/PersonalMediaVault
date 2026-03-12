// Common request errors composable

"use strict";

import type { Ref } from "vue";
import { ref } from "vue";
import { useI18n } from "./use-i18n";
import { emitAppEvent, EVENT_NAME_UNAUTHORIZED } from "@/global-state/app-events";
import { AuthController } from "@/global-state/auth";

/**
 * Composable that provides handlers for common errors
 */
export type CommonRequestErrorsComposable = {
    /**
     * Error message
     */
    error: Ref<string>;

    /**
     * True to display the error
     */
    errorDisplay: Ref<boolean>;

    /**
     * Sets the error message
     * @param errorMessage The error message
     */
    setError: (errorMessage: string) => void;

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
     * Handler for 'not found' error
     */
    notFound: () => void;

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

    // True to display the error
    const errorDisplay = ref(false);

    /**
     * Sets the error message
     * @param errorMessage The error message
     */
    const setError = (errorMessage: string) => {
        error.value = errorMessage;
        errorDisplay.value = !!errorMessage;
    };

    // Handler for 'unauthorized' error
    const unauthorized = () => {
        setError($t("Your session token expired") + ". " + $t("Please re-login in order to continue") + ".");
        emitAppEvent(EVENT_NAME_UNAUTHORIZED);
    };

    // Handler for 'bad request' error
    const badRequest = () => {
        setError(
            $t("The server unexpectedly rejected the request") +
                " " +
                $t("This may indicate you are using an outdated client version") +
                ". " +
                $t("Try refreshing the page") +
                ".",
        );
    };

    // Handler for 'access denied' error
    const accessDenied = () => {
        setError($t("Access denied") + ". " + $t("Try refreshing the page to refresh your permissions") + ".");
        AuthController.CheckAuthStatusSilent();
    };

    // Handler for 'not found' error
    const notFound = () => {
        setError($t("The target resource was not found") + ". " + $t("Try refreshing the page to re-synchronize with the server") + ".");
    };

    // Handler for internal server errors
    const serverError = () => {
        setError($t("Internal server error") + ".");
    };

    // Handler for network errors
    const networkError = () => {
        setError($t("Could not connect to the server") + ".");
    };

    return {
        error,
        errorDisplay,
        setError,
        unauthorized,
        badRequest,
        accessDenied,
        notFound,
        serverError,
        networkError,
    };
}
