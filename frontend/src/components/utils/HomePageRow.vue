<template>
    <div
        class="home-page-row"
        :class="{ moving: moving, 'moving-over': movingOver, 'moving-self': movingSelf }"
        :style="{ '--actual-row-size': rowSize + '', '--row-scroll-index': rowIndex + '', top: movingTop, left: movingLeft }"
        draggable="true"
        tabindex="-1"
        @dragstart="onDrag"
    >
        <div class="home-page-row-inner">
            <div class="home-page-row-head">
                <div class="home-page-row-title" :title="getGroupName(group)">{{ getGroupName(group) }}</div>
                <div v-if="editing" class="home-page-row-head-buttons">
                    <button type="button" class="page-header-btn" :title="$t('Rename row')" @click="renameRow">
                        <i class="fas fa-pencil-alt"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Move row')" @click="moveRow">
                        <i class="fas fa-arrows-up-down-left-right"></i>
                    </button>
                    <button type="button" class="page-header-btn" :title="$t('Delete row')" @click="deleteRow">
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
                <div class="home-page-row-message">{{ $t("This row cannot be customized") }}</div>
            </div>

            <div
                v-else-if="!editing && !loadDisplay && firstLoaded && elements.length === 0"
                class="home-page-row-content home-page-row-loading"
            >
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
                <div class="home-page-row-message">{{ $t("This row is empty") }}</div>
            </div>

            <div v-else-if="loading && !firstLoaded" class="home-page-row-content home-page-row-loading">
                <div v-for="f in loadingFiller" :key="f" class="search-result-item" :class="{ hidden: !loadDisplay }">
                    <div class="search-result-thumb">
                        <div class="search-result-thumb-inner">
                            <div class="search-result-loader">
                                <i class="fa fa-spinner fa-spin"></i>
                            </div>
                        </div>
                    </div>
                    <div v-if="displayTitles" class="search-result-title">{{ $t("Loading") }}...</div>
                </div>
            </div>

            <div v-else class="home-page-row-content"></div>
        </div>
    </div>
</template>

<script lang="ts">
import type { HomePageElement, HomePageGroup } from "@/api/api-home";
import { HomePageGroupTypes } from "@/api/api-home";
import type { HomePageGroupStartMovingData } from "@/utils/home";
import { getDefaultGroupName } from "@/utils/home";
import { clearNamedTimeout } from "@/utils/named-timeouts";
import { isTouchDevice } from "@/utils/touch";
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

        moving: Boolean,
        movingOver: Boolean,
        movingSelf: Boolean,
        movingLeft: String,
        movingTop: String,

        group: Object as PropType<HomePageGroup>,

        loadTick: Number,
    },
    emits: ["request-rename", "request-move", "request-delete", "start-moving"],
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

            loadDisplay: false,

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

            return this.getDefaultGroupName(group.type, this.$t);
        },

        getDefaultGroupName: getDefaultGroupName,

        renameRow: function () {
            this.$emit("request-rename", this.group);
        },

        moveRow: function () {
            this.$emit("request-move", this.group);
        },

        deleteRow: function () {
            this.$emit("request-delete", this.group);
        },

        onDrag: function (event: DragEvent) {
            event.preventDefault();

            if (!this.editing) {
                return;
            }

            if (isTouchDevice()) {
                return;
            }

            const startX = event.pageX;
            const startY = event.pageY;

            const bounds = (this.$el as HTMLElement).getBoundingClientRect();

            const data: HomePageGroupStartMovingData = {
                startX,
                startY,
                offsetX: startX - bounds.left,
                offsetY: startY - bounds.top,
                width: bounds.width,
                height: bounds.height,
            };

            this.$emit("start-moving", this.group, data);
        },
    },
});
</script>
