<template>
    <div class="thumb-bottom-right-tag" :title="renderTitle(type, duration)" :class="{ small: !!small }">
        <i v-if="type === 3" class="fas fa-headphones"></i>
        {{ renderTime(duration) }}
    </div>
</template>

<script lang="ts">
import { renderTimeSeconds } from "@/utils/time";
import { defineComponent } from "vue";

export default defineComponent({
    name: "DurationIndicator",
    props: {
        duration: Number,
        type: Number,
        small: Boolean,
    },
    methods: {
        renderTitle: function (type: number, duration: number): string {
            const typeTagName = type === 3 ? this.$t("Audio") : this.$t("Video");
            return typeTagName + " (" + this.renderTime(duration) + ")";
        },

        renderTime: function (s: number): string {
            return renderTimeSeconds(s);
        },
    },
});
</script>
