<template>
    <div class="paginated-menu" :class="{ 'menu-min': min }">
        <a
            target="_blank"
            :href="getPageUrl(0, pageName, order)"
            rel="noopener noreferrer"
            class="paginated-menu-link paginated-menu-link-edge-extra"
            @click="preventDefaultEvent"
        >
            <button
                :disabled="page <= 0"
                type="button"
                class="paginated-menu-btn paginated-menu-btn-edge-extra"
                :title="$t('First page')"
                @click="clickPage(0, $event)"
            >
                <i class="fas fa-angles-left"></i>
            </button>
        </a>

        <a
            target="_blank"
            :href="getPageUrl(Math.max(0, page - 1), pageName, order)"
            rel="noopener noreferrer"
            class="paginated-menu-link paginated-menu-link-edge"
            @click="preventDefaultEvent"
        >
            <button
                :disabled="page <= 0"
                type="button"
                class="paginated-menu-btn paginated-menu-btn-edge"
                :title="$t('Previous page')"
                @click="clickPage(page - 1, $event)"
            >
                <i class="fas fa-chevron-left"></i>
            </button>
        </a>

        <a
            v-for="(m, i) in menu"
            :key="i"
            target="_blank"
            :href="m.type === 'skip' ? '#' : getPageUrl(m.page, pageName, order)"
            rel="noopener noreferrer"
            class="paginated-menu-link"
            :class="{ current: m.current, skip: m.type === 'skip' }"
            @click="preventDefaultEvent"
        >
            <button
                type="button"
                class="paginated-menu-btn"
                :class="{ current: m.current, skip: m.type === 'skip' }"
                :disabled="m.type === 'skip'"
                @click="clickPage(m.page, $event)"
            >
                {{ m.type === "skip" ? "..." : m.page + 1 }}
            </button>
        </a>

        <a
            target="_blank"
            :href="getPageUrl(Math.min(page + 1, pages - 1), pageName, order)"
            rel="noopener noreferrer"
            class="paginated-menu-link paginated-menu-link-edge"
            @click="preventDefaultEvent"
        >
            <button
                :disabled="page >= pages - 1"
                type="button"
                class="paginated-menu-btn paginated-menu-btn-edge"
                :title="$t('Next page')"
                @click="clickPage(page + 1, $event)"
            >
                <i class="fas fa-chevron-right"></i>
            </button>
        </a>

        <a
            target="_blank"
            :href="getPageUrl(pages - 1, pageName, order)"
            rel="noopener noreferrer"
            class="paginated-menu-link paginated-menu-link-edge-extra"
            @click="preventDefaultEvent"
        >
            <button
                :disabled="page >= pages - 1"
                type="button"
                class="paginated-menu-btn paginated-menu-btn-edge-extra"
                :title="$t('Last page')"
                @click="clickPage(pages - 1, $event)"
            >
                <i class="fas fa-angles-right"></i>
            </button>
        </a>
    </div>
</template>

<script setup lang="ts">
import { useI18n } from "@/composables/use-i18n";
import type { AppStatusPage } from "@/control/app-status";

import { getFrontendUrl } from "@/utils/api";
import { preventDefaultEvent } from "@/utils/events";
import type { PageNode } from "@/utils/menu-make";
import { generateMenuForPages } from "@/utils/menu-make";
import { packSearchParams } from "@/utils/search-params";

import { ref, watch, type PropType } from "vue";

const { $t } = useI18n();

const emit = defineEmits<{
    /**
     * Event emitted when the user click a page
     */
    (e: "goto", page: number): void;
}>();

const props = defineProps({
    /**
     * Name of the page
     */
    pageName: {
        type: String as PropType<AppStatusPage>,
        required: true,
    },

    /**
     * The page current order
     */
    order: String,

    /**
     * The current page number
     */
    page: {
        type: Number,
        required: true,
    },

    /**
     * The max number of pages
     */
    pages: {
        type: Number,
        required: true,
    },

    /**
     * True to use miniature style
     */
    min: Boolean,
});

/// The menu as an array of page nodes
const menu = ref<PageNode[]>(generateMenuForPages(props.page, props.pages));

watch([() => props.page, () => props.pages], () => {
    menu.value = generateMenuForPages(props.page, props.pages);
});

/**
 * Called when a page button is clicked
 * @param p The page number
 * @param e The click event
 */
const clickPage = (p: number, e: Event) => {
    e.preventDefault();
    emit("goto", p);
};

/**
 * Gets full URL to a page
 * @param page The page number
 * @param pageName The page name
 * @param order The page order
 * @returns The URL
 */
const getPageUrl = (page: number, pageName: AppStatusPage, order: string): string => {
    return getFrontendUrl({
        page: pageName || null,
        sp: packSearchParams(page, order || "") || null,
    });
};
</script>
