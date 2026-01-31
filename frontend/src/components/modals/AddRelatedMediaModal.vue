<template>
    <ModalDialogContainer v-model:display="displayStatus" :close-signal="closeSignal" :static="true" @scroll.passive="onPageScroll">
        <div
            v-if="display"
            class="modal-dialog modal-xl modal-height-100"
            role="document"
            :class="{ 'rounded-corners-cells': roundedCorners }"
            :style="{
                '--row-size': rowSize,
                '--row-size-min': rowSizeMin,
                '--min-cell-size': minItemSize + 'px',
                '--max-cell-size': maxItemSize + 'px',
                '--cell-padding': padding + 'px',
            }"
        >
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Add related media") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding">
                <PageSearch
                    :display="true"
                    :in-modal="true"
                    :min="false"
                    :no-album="-1"
                    :page-size="pageSize"
                    :display-titles="displayTitles"
                    :row-size="rowSize"
                    :row-size-min="rowSizeMin"
                    :min-items-size="minItemSize"
                    :max-items-size="maxItemSize"
                    :remove-media-from-list="mediaElements"
                    @select-media="selectMedia"
                ></PageSearch>
            </div>

            <div v-if="pageScroll > 0" class="modal-button-br-container">
                <button type="button" :title="$t('Go to the top')" class="modal-button-br" @click="goTop">
                    <i class="fas fa-angles-up"></i>
                </button>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

import PageSearch from "@/components/pages/PageSearch.vue";
import {
    emitAppEvent,
    EVENT_NAME_ADVANCED_SEARCH_GO_TOP,
    EVENT_NAME_ADVANCED_SEARCH_SCROLL,
    EVENT_NAME_PAGE_PREFERENCES_UPDATED,
} from "@/control/app-events";
import { getPagePreferences } from "@/control/app-preferences";
import type { MediaListItem } from "@/api/models";

export default defineComponent({
    name: "AddRelatedMediaModal",
    components: {
        PageSearch,
    },
    props: {
        display: Boolean,
        mid: Number,
        relatedMedia: Array as PropType<MediaListItem[]>,
    },
    emits: ["update:display", "add-media"],
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        const pagePreferences = getPagePreferences("search");
        return {
            busy: false,

            pageSize: pagePreferences.pageSize,

            rowSize: pagePreferences.rowSize,
            rowSizeMin: pagePreferences.rowSizeMin,

            minItemSize: pagePreferences.minItemSize,
            maxItemSize: pagePreferences.maxItemSize,

            padding: pagePreferences.padding,

            displayTitles: pagePreferences.displayTitles,
            roundedCorners: pagePreferences.roundedCorners,

            pageScroll: 0,

            closeSignal: 0,

            mediaElements: new Set<number>((this.relatedMedia || []).map((e) => e.id).concat(this.mid)),
        };
    },
    watch: {
        display: function () {
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
            }
        },
        mid: function () {
            this.mediaElements = new Set<number>((this.relatedMedia || []).map((e) => e.id).concat(this.mid));
        },
        groupElements: function () {
            this.mediaElements = new Set<number>((this.relatedMedia || []).map((e) => e.id).concat(this.mid));
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_PAGE_PREFERENCES_UPDATED, this.updatePagePreferences.bind(this));

        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    methods: {
        close: function () {
            this.closeSignal++;
        },

        selectMedia: function (m: MediaListItem, callback: () => void) {
            this.$emit("add-media", m, callback);
        },

        updatePagePreferences: function () {
            const pagePreferences = getPagePreferences("search");

            this.pageSize = pagePreferences.pageSize;

            this.rowSize = pagePreferences.rowSize;
            this.rowSizeMin = pagePreferences.rowSizeMin;

            this.minItemSize = pagePreferences.minItemSize;
            this.maxItemSize = pagePreferences.maxItemSize;

            this.padding = pagePreferences.padding;

            this.displayTitles = pagePreferences.displayTitles;
            this.roundedCorners = pagePreferences.roundedCorners;
        },

        onPageScroll: function (e: Event) {
            this.pageScroll = (e.target as HTMLElement).scrollTop;

            emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_SCROLL, e);
        },

        goTop: function () {
            emitAppEvent(EVENT_NAME_ADVANCED_SEARCH_GO_TOP);

            nextTick(() => {
                this.$el.scrollTop = 0;
                this.$el.focus();
            });
        },
    },
});
</script>
