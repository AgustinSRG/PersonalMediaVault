<template>
    <div class="home-page-row" :style="{ '--actual-row-size': rowSize + '', '--row-scroll-index': rowIndex }">
        <div class="home-page-row-head">
            <div class="home-page-row-title" :title="getGroupName(group)">{{ getGroupName(group) }}</div>
            <div v-if="editing" class="home-page-row-head-buttons">
                <button type="button" class="page-header-btn" :title="$t('Rename row')">
                    <i class="fas fa-pencil-alt"></i>
                </button>
                <button type="button" class="page-header-btn" :title="$t('Move row')">
                    <i class="fas fa-arrows-up-down-left-right"></i>
                </button>
                <button type="button" class="page-header-btn" :title="$t('Delete row')">
                    <i class="fas fa-trash-alt"></i>
                </button>
            </div>
        </div>

        <div v-if="editing && group.type != groupTypeCustom" class="home-page-row-content home-page-row-loading">
            <div v-for="f in loadingFiller.slice(1)" :key="f" class="search-result-item hidden">
                <div class="search-result-thumb">
                    <div class="search-result-thumb-inner">
                        <div class="search-result-loader">
                            <i class="fa fa-spinner fa-spin"></i>
                        </div>
                    </div>
                </div>
                <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
            </div>
            <div class="home-page-row-message">{{ $t("This row cannot be customized") }} ({{ getDefaultGroupName(group) }})</div>
        </div>

        <div v-else-if="loading || elements.length === 0" class="home-page-row-content home-page-row-loading">
            <div
                v-for="f in loadingFiller"
                :key="f"
                class="search-result-item"
                :class="{ hidden: elements.length === 0 && !loading && firstLoaded }"
            >
                <div class="search-result-thumb">
                    <div class="search-result-thumb-inner">
                        <div class="search-result-loader">
                            <i class="fa fa-spinner fa-spin"></i>
                        </div>
                    </div>
                </div>
                <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
            </div>
            <div v-if="elements.length === 0 && !loading && firstLoaded" class="home-page-row-message">
                <i class="fas fa-box-open"></i> {{ $t("This row is empty") }}
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import type { HomePageElement, HomePageGroup } from "@/api/api-home";
import { HomePageGroupTypes } from "@/api/api-home";
import { clearNamedTimeout } from "@/utils/named-timeouts";
import { getUniqueStringId } from "@/utils/unique-id";
import { abortNamedApiRequest } from "@asanrom/request-browser";
import type { PropType } from "vue";
import { defineComponent } from "vue";

export default defineComponent({
    name: "HomePageRow",
    props: {
        rowSize: Number,

        pageSize: Number,

        displayTitles: Boolean,

        editing: Boolean,

        group: Object as PropType<HomePageGroup>,

        loadTick: Number,
    },
    setup() {
        return {
            groupTypeCustom: HomePageGroupTypes.Custom,

            loadRequestId: getUniqueStringId(),
        };
    },
    data: function () {
        return {
            rowIndex: 0,
            rowSplitCount: 1,

            loading: true,
            firstLoaded: false,

            elements: [] as HomePageElement[],

            loadingFiller: Array(this.pageSize)
                .fill(0)
                .map((_v, i) => i),
        };
    },
    watch: {
        pageSize: function () {
            this.loadingFiller = Array(this.pageSize)
                .fill(0)
                .map((_v, i) => i);
        },
    },
    mounted: function () {},
    beforeUnmount: function () {
        clearNamedTimeout(this.loadRequestId);
        abortNamedApiRequest(this.loadRequestId);
    },
    methods: {
        getGroupName(group: HomePageGroup): string {
            if (group.name) {
                return group.name;
            }

            return this.getDefaultGroupName(group);
        },

        getDefaultGroupName(group: HomePageGroup): string {
            switch (group.type) {
                case HomePageGroupTypes.RecentMedia:
                    return this.$t("Media");
                case HomePageGroupTypes.RecentAlbums:
                    return this.$t("Albums");
                default:
                    return this.$t("Custom row");
            }
        },
    },
});
</script>
