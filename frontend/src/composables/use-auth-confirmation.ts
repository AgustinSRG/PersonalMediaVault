// Auth confirmation composable

"use strict";

import type { Ref } from "vue";
import { ref } from "vue";
import { useI18n } from "./use-i18n";

/**
 * Composable for auth confirmation
 */
export type AuthConfirmationComposable = {
    /**
     * Auth confirmation modal display
     */
    displayAuthConfirmation: Ref<boolean>;

    /**
     * Auth confirmation cooldown
     */
    authConfirmationCooldown: Ref<number>;

    /**
     * True if auth confirmation requires TFA,
     * false if it requires password
     */
    authConfirmationTfa: Ref<boolean>;

    /**
     * Auth confirmation error.
     */
    authConfirmationError: Ref<string>;

    /**
     * Call when a request requires auth confirmation
     * with password mode.
     */
    requiredAuthConfirmationPassword: () => void;

    /**
     * Call when an auth confirmation request
     * fails because the password is invalid
     */
    invalidPassword: () => void;

    /**
     * Call when a request requires auth confirmation
     * with TFA mode.
     */
    requiredAuthConfirmationTfa: () => void;

    /**
     * Call when an auth confirmation request
     * fails because the TFA code is invalid
     */
    invalidTfaCode: () => void;

    /**
     * Call when an auth confirmation request
     * fails because of cooldown
     */
    cooldown: () => void;
};

// Auth confirmation cooldown (milliseconds)
const AUTH_CONFIRMATION_COOLDOWN = 5000;

/**
 * Creates necessary references for AuthConfirmationModal
 * @returns The references in order to use AuthConfirmationModal
 */
export function useAuthConfirmation(): AuthConfirmationComposable {
    // Translation function
    const { $t } = useI18n();

    // Modal display
    const displayAuthConfirmation = ref(false);

    // Cooldown
    const authConfirmationCooldown = ref(0);

    // TFA mode
    const authConfirmationTfa = ref(false);

    // Error message
    const authConfirmationError = ref("");

    // Call when a request requires auth confirmation
    // with password mode.
    const requiredAuthConfirmationPassword = () => {
        displayAuthConfirmation.value = true;
        authConfirmationError.value = "";
        authConfirmationTfa.value = false;
    };

    // Call when an auth confirmation request
    // fails because the password is invalid
    const invalidPassword = () => {
        displayAuthConfirmation.value = true;
        authConfirmationError.value = $t("Invalid password");
        authConfirmationTfa.value = false;
        authConfirmationCooldown.value = Date.now() + AUTH_CONFIRMATION_COOLDOWN;
    };

    // Call when a request requires auth confirmation
    // with TFA mode.
    const requiredAuthConfirmationTfa = () => {
        displayAuthConfirmation.value = true;
        authConfirmationError.value = "";
        authConfirmationTfa.value = true;
    };

    // Call when an auth confirmation request
    // fails because the TFA code is invalid
    const invalidTfaCode = () => {
        displayAuthConfirmation.value = true;
        authConfirmationError.value = $t("Invalid one-time code");
        authConfirmationTfa.value = true;
        authConfirmationCooldown.value = Date.now() + AUTH_CONFIRMATION_COOLDOWN;
    };

    // Call when an auth confirmation request
    // fails because of cooldown
    const cooldown = () => {
        displayAuthConfirmation.value = true;
        authConfirmationError.value = $t("You must wait 5 seconds to try again");
    };

    return {
        displayAuthConfirmation,
        authConfirmationCooldown,
        authConfirmationTfa,
        authConfirmationError,
        requiredAuthConfirmationPassword,
        invalidPassword,
        requiredAuthConfirmationTfa,
        invalidTfaCode,
        cooldown,
    };
}
