// Composable for I18N

"use strict";

import { localeRef, translate } from "@/i18n";
import type { Ref } from "vue";
import { computed } from "vue";

/**
 * I18N composable
 */
export type ComposableI18N = {
    /**
     * The current locale
     */
    locale: Ref<string>;

    /**
     * The translation function
     * @param key The translation key
     * @returns The translated value
     */
    $t: (key: string) => string;
};

/**
 * Gets the locale and the translation function
 * @returns The locale and the translation function
 */
export function useI18n(): ComposableI18N {
    const translateFn = computed(() => {
        return (key: string) => {
            return translate(key, localeRef.value);
        };
    });

    return {
        locale: localeRef,
        $t: translateFn.value,
    };
}
