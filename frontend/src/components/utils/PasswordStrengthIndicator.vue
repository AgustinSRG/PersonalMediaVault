<template>
    <div class="password-strength-indicator">
        <div
            class="password-strength-indicator-progress"
            :style="{ '--segment-count': segments.length }"
            :class="{
                [strengthTag]: true,
            }"
        >
            <div
                v-for="s in segments"
                :key="s"
                class="password-strength-indicator-progress-segment"
                :class="{ active: isSegmentActive(s, strength) }"
            ></div>
        </div>
        <div class="password-strength-indicator-label">
            {{ $t("Password strength") }}: {{ renderPasswordStrengthTag(strengthTag) }} ({{ renderPasswordStrength(strength) }})
        </div>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import type { PasswordStrengthTag } from "@/utils/password-strength";
import { computePasswordStrength, getPasswordStrengthTag } from "@/utils/password-strength";
import { computed } from "vue";

const { $t } = useI18n();

/// Min entropy for a password to be considered cryptographically secure
const CRYPTO_SECURE_ENTROPY = 128;

/// Number of segments for the strength indicator
const STRENGTH_SEGMENTS = 16;

const props = defineProps({
    /**
     * The password to check
     */
    password: String,
});

/**
 * Creates an array of segment indexes
 * @param n The number of segments
 * @return The array of segment indexes
 */
function createSegments(n: number): number[] {
    const segments: number[] = [];

    for (let i = 0; i < n; i++) {
        segments.push(i);
    }

    return segments;
}

/// Array of segment indexes
const segments = createSegments(STRENGTH_SEGMENTS);

/**
 * Checks if a segment is active based on the password strength
 * @param segmentIndex The segment index
 * @param strength The password strength
 * @returns True if the segment is active
 */
const isSegmentActive = (segmentIndex: number, strength: number): boolean => {
    const segmentVal = (CRYPTO_SECURE_ENTROPY / STRENGTH_SEGMENTS) * segmentIndex;
    return strength >= segmentVal;
};

/// Password strength
const strength = computed(() => computePasswordStrength(props.password || ""));

/// Password strength tag
const strengthTag = computed(() => getPasswordStrengthTag(strength.value));

/**
 * Renders password strength as a number with 1 decimal place
 * @param s The password strength
 * @return The rendered strength
 */
const renderPasswordStrength = (s: number): string => {
    return (Math.round(s * 10) / 10).toFixed(1);
};

/**
 * Renders password strength tag
 * @param tag Th password strength tag
 * @returns The rendered tag
 */
const renderPasswordStrengthTag = (tag: PasswordStrengthTag): string => {
    switch (tag) {
        case "very-weak":
            return $t("Very weak");
        case "weak":
            return $t("Weak");
        case "medium":
            return $t("Medium");
        case "strong":
            return $t("Strong");
        case "very-strong":
            return $t("Very strong");
        case "crypto-secure":
            return $t("Cryptographically secure");
        default:
            return "???";
    }
};
</script>
