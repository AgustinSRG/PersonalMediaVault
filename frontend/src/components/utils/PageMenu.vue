<template>
  <div class="paginated-menu" :class="{ 'menu-min': min }">
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

<style>
.paginated-menu {
  display: flex;
  flex-direction: row;
  justify-content: center;
  padding: 0.5rem;
}

.paginated-menu-btn {
  display: block;
  padding: 0.5rem;
  min-width: 40px;
  min-height: 40px;
  box-shadow: none;
  border: none;
  border-radius: 100vw;
  cursor: pointer;
  font-size: 18px;
  font-weight: bold;
  background: transparent;
  margin: 0.5rem;
  color: white;
}

.paginated-menu-btn:disabled {
  opacity: 0.7;
  cursor: default;
}

.paginated-menu-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}

.paginated-menu-btn:disabled:hover {
  background: none;
}

.paginated-menu-btn.current,
.paginated-menu-btn.current:hover {
  color: black;
  background: white;
}

.menu-min .paginated-menu-btn {
  display: none;
}

.menu-min .paginated-menu-btn.current,
.menu-min .paginated-menu-btn.paginated-menu-btn-edge {
  display: block;
}

@media (max-width: 1000px) {
  .paginated-menu-btn {
    display: none;
  }

  .paginated-menu-btn.current,
  .paginated-menu-btn.paginated-menu-btn-edge {
    display: block;
  }
}
</style>