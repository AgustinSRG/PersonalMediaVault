<template>
    <div class="paginated-menu" :class="{ 'menu-min': min }">
        <button
            :disabled="page <= 0"
            type="button"
            class="paginated-menu-btn paginated-menu-btn-edge-extra"
            @click="clickPage(0)"
            :title="$t('First page')"
        >
            <i class="fas fa-angles-left"></i>
        </button>

        <button
            :disabled="page <= 0"
            type="button"
            class="paginated-menu-btn paginated-menu-btn-edge"
            @click="clickPage(page - 1)"
            :title="$t('Previous page')"
        >
            <i class="fas fa-chevron-left"></i>
        </button>

        <button
            v-for="(m, i) in menu"
            :key="i"
            type="button"
            class="paginated-menu-btn"
            :class="{ current: m.current, skip: m.type === 'skip' }"
            :disabled="m.type === 'skip'"
            @click="clickPage(m.page)"
        >
            {{ m.type === "skip" ? "..." : m.page + 1 }}
        </button>

        <button
            :disabled="page >= pages - 1"
            type="button"
            class="paginated-menu-btn paginated-menu-btn-edge"
            @click="clickPage(page + 1)"
            :title="$t('Next page')"
        >
            <i class="fas fa-chevron-right"></i>
        </button>

        <button
            :disabled="page >= pages - 1"
            type="button"
            class="paginated-menu-btn paginated-menu-btn-edge-extra"
            @click="clickPage(pages - 1)"
            :title="$t('Last page')"
        >
            <i class="fas fa-angles-right"></i>
        </button>
    </div>
</template>

<script lang="ts">
import { generateMenuForPages } from "@/utils/menu-make";
import { defineComponent } from "vue";

export default defineComponent({
    name: "PageMenu",
    emits: ["goto"],
    props: {
        page: Number,
        pages: Number,
        min: Boolean,
    },
    data: function () {
        return {
            menu: [],
        };
    },
    methods: {
        updatePageMenu: function () {
            this.menu = generateMenuForPages(this.page, this.pages);
        },

        clickPage: function (p) {
            this.$emit("goto", p);
        },
    },
    mounted: function () {
        this.updatePageMenu();
    },
    watch: {
        page: function () {
            this.updatePageMenu();
        },
        pages: function () {
            this.updatePageMenu();
        },
    },
});
</script>
