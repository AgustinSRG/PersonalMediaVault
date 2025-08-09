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

<script lang="ts">
import type { PasswordStrengthTag } from "@/utils/password-strength";
import { computePasswordStrength, getPasswordStrengthTag } from "@/utils/password-strength";
import { defineComponent } from "vue";

const CRYPTO_SECURE_ENTROPY = 128;
const STRENGTH_SEGMENTS = 16;

export default defineComponent({
    name: "PasswordStrengthIndicator",
    props: {
        password: String,
    },
    setup: function () {
        const segments: number[] = [];

        for (let i = 0; i < STRENGTH_SEGMENTS; i++) {
            segments.push(i);
        }

        return {
            segments,
        };
    },
    data: function () {
        const passwordStrength = computePasswordStrength(this.password || "");
        return {
            strength: passwordStrength,
            strengthTag: getPasswordStrengthTag(passwordStrength),
        };
    },
    watch: {
        password: function () {
            this.strength = computePasswordStrength(this.password || "");
            this.strengthTag = getPasswordStrengthTag(this.strength);
        },
    },
    mounted: function () {},
    beforeUnmount: function () {},
    methods: {
        isSegmentActive: function (s: number, strength: number): boolean {
            const segmentVal = (CRYPTO_SECURE_ENTROPY / STRENGTH_SEGMENTS) * s;
            return strength >= segmentVal;
        },

        renderPasswordStrength: function (s: number) {
            return (Math.round(s * 10) / 10).toFixed(1);
        },

        renderPasswordStrengthTag: function (tag: PasswordStrengthTag): string {
            switch (tag) {
                case "very-weak":
                    return this.$t("Very weak");
                case "weak":
                    return this.$t("Weak");
                case "medium":
                    return this.$t("Medium");
                case "strong":
                    return this.$t("Strong");
                case "very-strong":
                    return this.$t("Very strong");
                case "crypto-secure":
                    return this.$t("Cryptographically secure");
                default:
                    return "???";
            }
        },
    },
});
</script>
