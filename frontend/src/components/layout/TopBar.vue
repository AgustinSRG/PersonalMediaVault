<template>
  <div class="top-bar" tabindex="-1">
    <div class="top-bar-logo-td">
      <button type="button" class="top-bar-button top-bar-menu-btn" :title="$t('Main menu')" @click="menu">
        <i class="fas fa-bars"></i>
      </button>
      <img class="top-bar-logo-img" src="@/assets/favicon.png" alt="PMV" />
      <span :title="getAppTitle()" class="top-bar-title">PMV</span>
    </div>
    <div class="top-bar-search-td">
      <div class="top-bar-center-div">
        <form class="top-bar-search-input-container" :class="{ focused: searchFocus }" @submit="submitSearch" tabindex="-1">
          <input type="text" class="top-bar-search-input" name="pmv-search-input" spellcheck="false" autocorrect="off" autocomplete="off" autocapitalize="none" :placeholder="$t('Search')" v-model="search" @keydown="onKeyDown" @input="onSearchInput" @focus="focusSearch" />
          <button type="submit" class="top-bar-button top-bar-search-button" :title="$t('Search')" @focus="blurSearch">
            <i class="fas fa-search"></i>
          </button>
          <div class="top-bar-search-suggestions" :class="{ hidden: suggestions.length === 0 }">
            <div v-for="s in suggestions" :key="s.key" class="top-bar-search-suggestion" @click="clickSearch(s)">
              <i class="fas fa-tag" v-if="s.type === 'tag'"></i>
              <i class="fas fa-list-ol" v-else-if="s.type === 'album'"></i>
              <span>{{ s.name }}</span>
            </div>
          </div>
        </form>
      </div>
    </div>
    <div class="top-bar-user-td">
      <button type="button" class="top-bar-button top-bar-button-dropdown top-bar-button-large-version" :title="$t('Help')" @click="help">
        <i class="fas fa-question"></i>
      </button>

      <button type="button" class="top-bar-button top-bar-button-small-version" :title="$t('Search')" @click="openSearch">
        <i class="fas fa-search"></i>
      </button>

      <button type="button" class="top-bar-button top-bar-button-dropdown" :title="$t('Settings')" @click="settings">
        <i class="fas fa-cog"></i>
      </button>
      <button type="button" class="top-bar-button" :title="$t('Close vault')" @click="logout">
        <i class="fas fa-sign-out-alt"></i>
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { KeyboardManager } from "@/control/keyboard";
import { TagsController } from "@/control/tags";
import { defineComponent } from "vue";
import { FocusTrap } from "../../utils/focus-trap";

export default defineComponent({
  name: "TopBar",
  emits: ["logout", "settings", "menu", "search-open", "help"],
  data: function () {
    return {
      search: AppStatus.CurrentSearch,
      searchFocus: false,
      suggestions: [],
    };
  },
  methods: {
    getAppTitle: function () {
      return AuthController.Title || this.$t("Personal Media Vault");
    },

    menu: function () {
      this.$emit("menu");
    },

    logout: function () {
      this.$emit("logout");
    },

    settings: function () {
      this.$emit("settings");
    },

    openSearch: function () {
      this.$emit("search-open");
    },

    help: function () {
      this.$emit("help");
    },

    goSearch: function () {
      AppStatus.GoToSearch(this.search);
      this.$el.querySelector(".top-bar-search-input").blur();
    },

    onSearchChanged: function () {
      this.search = AppStatus.CurrentSearch;
    },

    focusSearch: function () {
      if (this.$options.blurTimeout) {
        clearTimeout(this.$options.blurTimeout);
        this.$options.blurTimeout = null;
      }
      this.searchFocus = true;
      this.$el.querySelector(".top-bar-search-input").select();
      this.updateSuggestions();
      if (this.$options.focusTrap) {
        this.$options.focusTrap.activate();
      }
    },

    blurSearchInstantly: function () {
      if (this.$options.blurTimeout) {
        clearTimeout(this.$options.blurTimeout);
        this.$options.blurTimeout = null;
      }
      this.$options.blurTimeout = null;
      this.searchFocus = false;
      if (this.$options.focusTrap) {
        this.$options.focusTrap.deactivate();
      }
    },

    blurSearch: function () {
      if (this.$options.blurTimeout) {
        clearTimeout(this.$options.blurTimeout);
        this.$options.blurTimeout = null;
      }
      this.$options.blurTimeout = setTimeout(() => {
        this.$options.blurTimeout = null;
        this.searchFocus = false;
        if (this.$options.focusTrap) {
          this.$options.focusTrap.deactivate();
        }
      }, 100);
    },

    clickSearch: function (s) {
      this.blurSearch();
      this.selectSearch(s);
    },

    selectSearch: function (s) {
      if (s.type === "album") {
        this.search = "";
        AppStatus.ClickOnAlbum(s.id);
      } else {
        this.search = s.name;
      }
      this.goSearch();
      this.blurSearchInstantly();
    },

    onSearchModalSubmit: function (search: string) {
      this.search = search;
      this.submitSearch();
    },

    submitSearch: function (event) {
      if (event) {
        event.preventDefault();
      }
      this.blurSearchInstantly();
      if (!this.search) {
        this.goSearch();
        return;
      }
      this.updateSuggestions();
      if (this.suggestions.length > 0) {
        this.selectSearch(this.suggestions[0]);
      } else {
        this.goSearch();
      }
    },

    updateSuggestions: function () {
      const tagFilter = this.search
        .replace(/[\n\r]/g, " ")
        .trim()
        .replace(/[\s]/g, "_")
        .toLowerCase();
      this.suggestions = Object.values(TagsController.Tags)
        .map((a: any) => {
          const i = tagFilter ? a.name.indexOf(tagFilter) : 0;
          return {
            key: "tag:" + a.id,
            id: a.id,
            name: a.name,
            nameLower: a.name,
            starts: i === 0,
            contains: i >= 0,
            type: "tag",
          };
        })
        .concat(
          Object.values(AlbumsController.Albums).map((a) => {
            const name = a.name.replace(/[\s]/g, "_").toLowerCase();
            const i = tagFilter ? name.indexOf(tagFilter) : 0;
            return {
              key: "album:" + a.id,
              id: a.id,
              name: a.name,
              nameLower: name,
              starts: i === 0,
              contains: i >= 0,
              type: "album",
            };
          })
        )
        .filter((a) => {
          return a.starts || a.contains;
        })
        .sort((a, b) => {
          if (a.starts && !b.starts) {
            return -1;
          } else if (b.starts && !a.starts) {
            return 1;
          } else if (a.nameLower < b.nameLower) {
            return -1;
          } else {
            return 1;
          }
        });
    },

    onSearchInput: function () {
      if (this.$options.findTagTimeout) {
        return;
      }
      this.$options.findTagTimeout = setTimeout(() => {
        this.$options.findTagTimeout = null;
        this.updateSuggestions();
      }, 200);
    },

    onKeyDown: function (event) {
      if (event.key === "Tab" && this.search && !event.shiftKey) {
        if (
          this.suggestions.length > 0 &&
          this.search !== this.suggestions[0].name
        ) {
          this.search = this.suggestions[0].name;
          this.onSearchInput();
          event.preventDefault();
        }
      }
      event.stopPropagation();
    },

    handleGlobalKey: function (event: KeyboardEvent): boolean {
      if (AuthController.Locked || !event.key) {
        return false;
      }

      if (event.key.toUpperCase() === "M" && event.ctrlKey) {
        this.menu();
        return true;
      }

      if (event.key.toUpperCase() === "F" && event.ctrlKey) {
        const searchInput = this.$el.querySelector(".top-bar-search-input");
        if (searchInput) {
          searchInput.focus();
        }
        return true;
      }

      if (event.key.toUpperCase() === "S" && event.ctrlKey) {
        this.settings();
        return true;
      }

      if (event.key.toUpperCase() === "Q" && event.ctrlKey) {
        this.logout();
        return true;
      }

      return false;
    },
  },

  mounted: function () {
    this.$options.statusChangeH = this.onSearchChanged.bind(this);

    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    this.$options.onSearchModalSubmitH = this.onSearchModalSubmit.bind(this);
    AppEvents.AddEventListener(
      "search-modal-submit",
      this.$options.onSearchModalSubmitH
    );

    this.$options.handleGlobalKeyH = this.handleGlobalKey.bind(this);
    KeyboardManager.AddHandler(this.$options.handleGlobalKeyH);

    this.$options.focusTrap = new FocusTrap(
      this.$el.querySelector(".top-bar-search-input-container"),
      this.blurSearch.bind(this)
    );
  },

  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    AppEvents.RemoveEventListener(
      "search-modal-submit",
      this.$options.onSearchModalSubmitH
    );

    if (this.$options.findTagTimeout) {
      clearTimeout(this.$options.findTagTimeout);
    }

    if (this.$options.blurTimeout) {
      clearTimeout(this.$options.blurTimeout);
      this.$options.blurTimeout = null;
    }

    KeyboardManager.RemoveHandler(this.$options.handleGlobalKeyH);

    if (this.$options.focusTrap) {
      this.$options.focusTrap.destroy();
    }
  },
});

// Searchbox background: hsl(0, 0%, 7%)
// Searchbox text color : hsla(0, 100%, 100%, 0.88)
</script>
