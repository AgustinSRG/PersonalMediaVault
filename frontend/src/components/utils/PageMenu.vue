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

<script lang="ts">
import { generateURIQuery } from "@/utils/api";
import { generateMenuForPages } from "@/utils/menu-make";
import { packSearchParams } from "@/utils/search-params";
import { defineComponent } from "vue";

export default defineComponent({
    name: "PageMenu",
    props: {
        pageName: String,
        order: String,
        page: Number,
        pages: Number,
        min: Boolean,
    },
    emits: ["goto"],
    data: function () {
        return {
            menu: [],
        };
    },
    watch: {
        page: function () {
            this.updatePageMenu();
        },
        pages: function () {
            this.updatePageMenu();
        },
    },
    mounted: function () {
        this.updatePageMenu();
    },
    methods: {
        updatePageMenu: function () {
            this.menu = generateMenuForPages(this.page, this.pages);
        },

        clickPage: function (p: number, e: Event) {
            e.preventDefault();
            this.$emit("goto", p);
        },

        preventDefaultEvent: function (e: Event) {
            e.preventDefault();
        },

        getPageUrl: function (page: number, pageName: string, order: string): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                generateURIQuery({
                    page: pageName || "",
                    sp: packSearchParams(page, order || "") || null,
                })
            );
        },
    },
});
</script>
