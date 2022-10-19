<template>
  <div class="top-bar">
    <div class="top-bar-logo-td">
      <button
        type="button"
        class="top-bar-button top-bar-menu-btn"
        :title="$t('Main menu')"
        @click="menu"
      >
        <i class="fas fa-bars"></i>
      </button>
      <img class="top-bar-logo-img" src="@/assets/favicon.png" alt="PMV" />
      <span :title="$t('Personal Media Vault')" class="top-bar-title">PMV</span>
      <span :title="$t('Personal Media Vault')" class="top-bar-title-min"
        >PMV</span
      >
    </div>
    <div class="top-bar-search-td">
      <div class="top-bar-center-div">
        <form
          class="top-bar-search-input-container"
          :class="{ focused: searchFocus }"
          @submit="submitSearch"
        >
          <input
            type="text"
            class="top-bar-search-input"
            name="pmv-search-input"
            spellcheck="false"
            autocorrect="off"
            autocomplete="off"
            autocapitalize="none"
            :placeholder="$t('Search')"
            v-model="search"
            @keydown="onKeyDown"
            @input="onSearchInput"
            @focus="focusSearch"
            @blur="blurSearch"
          />
          <button
            type="submit"
            class="top-bar-button top-bar-search-button"
            :title="$t('Search')"
          >
            <i class="fas fa-search"></i>
          </button>
          <div
            class="top-bar-search-suggestions"
            :class="{ hidden: suggestions.length === 0 }"
          >
            <div
              v-for="s in suggestions"
              :key="s.key"
              class="top-bar-search-suggestion"
              @click="selectSearch(s)"
            >
              <i class="fas fa-tag" v-if="s.type === 'tag'"></i>
              <i class="fas fa-list-ol" v-else-if="s.type === 'album'"></i>
              <span>{{ s.name }}</span>
            </div>
          </div>
        </form>
      </div>
    </div>
    <div class="top-bar-user-td">
      <button
        type="button"
        class="top-bar-button top-bar-bottom-min-version"
        :title="$t('Search')"
      >
        <i class="fas fa-search"></i>
      </button>

      <button
        type="button"
        class="top-bar-button"
        :title="$t('Settings')"
        @click="settings"
      >
        <i class="fas fa-cog"></i>
      </button>
      <button
        type="button"
        class="top-bar-button"
        :title="$t('Close vault')"
        @click="logout"
      >
        <i class="fas fa-sign-out-alt"></i>
      </button>
    </div>
  </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { TagsController } from "@/control/tags";
import { defineComponent } from "vue";

export default defineComponent({
  name: "TopBar",
  emits: ["logout", "settings", "menu"],
  data: function () {
    return {
      search: AppStatus.CurrentSearch,
      searchFocus: false,
      suggestions: [],
    };
  },
  methods: {
    menu: function () {
      this.$emit("menu");
    },

    logout: function () {
      this.$emit("logout");
    },

    settings: function () {
      this.$emit("settings");
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
      this.updateSuggestions();
    },

    blurSearch: function () {
      if (this.$options.blurTimeout) {
        clearTimeout(this.$options.blurTimeout);
        this.$options.blurTimeout = null;
      }
      this.$options.blurTimeout = setTimeout(() => {
        this.$options.blurTimeout = null;
        this.searchFocus = false;
      }, 100);
    },

    selectSearch: function (s) {
      if (s.type === "album") {
        this.search = "";
        AppStatus.ClickOnAlbum(s.id);
      } else {
        this.search = s.name;
      }
      this.goSearch();
    },

    submitSearch: function (event) {
      if (event) {
        event.preventDefault();
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
      event.stopPropagation();
      if (event.key === "Tab" && this.search && !event.shiftKey) {
        if (this.suggestions.length > 0) {
          this.search = this.suggestions[0].name;
          this.onSearchInput();
          event.preventDefault();
        }
      }
    },
  },

  mounted: function () {
    this.$options.statusChangeH = this.onSearchChanged.bind(this);

    AppEvents.AddEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );
  },

  beforeUnmount: function () {
    AppEvents.RemoveEventListener(
      "app-status-update",
      this.$options.statusChangeH
    );

    if (this.$options.findTagTimeout) {
      clearTimeout(this.$options.findTagTimeout);
    }

    if (this.$options.blurTimeout) {
      clearTimeout(this.$options.blurTimeout);
      this.$options.blurTimeout = null;
    }
  },
});

// Searchbox background: hsl(0, 0%, 7%)
// Searchbox text color : hsla(0, 100%, 100%, 0.88)
</script>

<style>
.top-bar {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 56px;

  display: flex;
  justify-content: space-between;
  align-items: center;

  white-space: nowrap;
  z-index: 10;
}

.vault-locked .top-bar {
  visibility: hidden;
}

.light-theme .top-bar {
  background: #ffffff;
}

.dark-theme .top-bar {
  background: #212121;
}

.top-bar-logo-td {
  text-align: left;
  display: flex;
  align-items: center;
}

.top-bar-search-td {
  text-align: center;
}

.top-bar-user-td {
  padding-right: 8px;
  text-align: right;
}

.top-bar-title {
  font-weight: bold;
  font-size: 24px;
  padding-left: 8px;
  padding-top: 3px;
}

.top-bar-title-min {
  font-weight: bold;
  font-size: 24px;
  padding-left: 8px;
  padding-top: 3px;
  display: none;
}

@media (max-width: 1024px) {
  .top-bar-title {
    display: none;
  }

  .top-bar-title-min {
    display: initial;
  }
}

@media (max-width: 300px) {
  .top-bar-title-min {
    display: none;
  }
}

.top-bar-button {
  display: inline-block;
  width: 48px;
  height: 48px;
  box-shadow: none;
  border: none;
  cursor: pointer;
  font-size: 24px;
  background: transparent;
  color: var(--theme-btn-color);
}

.top-bar-button:disabled {
  opacity: 0.7;
  cursor: default;
}

.top-bar-button:not(:disabled):hover {
  color: var(--theme-btn-hover-color);
}

.top-bar-center-div {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.top-bar-search-input-container {
  height: 40px;
  display: inline-block;
  width: 480px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.dark-theme .top-bar-search-input-container {
  border: 1px solid hsl(0, 0%, 18.82%);
  background: hsla(0, 0%, 100%, 0.08);
}

.light-theme .top-bar-search-input-container {
  border: 1px solid hsl(0, 0%, 18.82%);
  background: white;
}

.light-theme .top-bar-search-input-container.focused {
  box-shadow: 0 0 0 0.1rem rgba(0, 0, 0, 0.1);
}

.dark-theme .top-bar-search-input-container.focused {
  box-shadow: 0 0 0 0.1rem rgba(255, 255, 255, 0.1);
}

.top-bar-search-input {
  outline: none;
  width: calc(100% - 54px);
  height: 38px;
  border: none;
  font-size: 16px;
  margin: 0;
  padding: 1px 4px;
}

.dark-theme .top-bar-search-input {
  color: white;
  background: hsl(0, 0%, 7%);
}

.light-theme .top-bar-search-input {
  color: black;
  background: white;
}

@media (max-width: 850px) {
  .top-bar-search-input-container {
    width: 360px;
  }
}

@media (max-width: 740px) {
  .top-bar-search-input-container {
    width: 240px;
  }
}

.top-bar-bottom-min-version {
  display: none;
}

@media (max-width: 600px) {
  .top-bar-search-input-container {
    display: none;
  }

  .top-bar-bottom-min-version {
    display: inline-block;
  }
}

.top-bar-logo-td::placeholder {
  color: hsla(0, 100%, 100%, 0.88);
}

.top-bar-search-button {
  width: 54px;
  height: 40px;
  font-size: 20px;
}

.top-bar-search-suggestions {
  position: absolute;
  top: calc(40px + 0.2rem);
  left: 0;
  width: calc(100% + 0.1rem);
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.1s, visibility 0.1s;

  display: flex;
  flex-direction: column;
  max-height: 300px;
  overflow-y: auto;
}

.light-theme .top-bar-search-suggestions {
  background: rgba(255, 255, 255, 0.8);
}

.dark-theme .top-bar-search-suggestions {
  background: rgba(0, 0, 0, 0.8);
}

.top-bar-search-input-container.focused .top-bar-search-suggestions {
  transition: opacity 0.1s;
  opacity: 1;
  pointer-events: all;
}

.light-theme
  .top-bar-search-input-container.focused
  .top-bar-search-suggestions {
  box-shadow: 0 0 0 0.1rem rgba(0, 0, 0, 0.1);
}

.dark-theme
  .top-bar-search-input-container.focused
  .top-bar-search-suggestions {
  box-shadow: 0 0 0 0.1rem rgba(255, 255, 255, 0.1);
}

.top-bar-search-input-container.focused .top-bar-search-suggestions.hidden {
  opacity: 0;
  pointer-events: none;
  visibility: hidden;
}

.top-bar-search-suggestion {
  width: 100%;
  padding: 0.5rem 1rem;
  font-size: 16px;
  cursor: pointer;
  text-align: left;
}

.light-theme .top-bar-search-suggestion:hover {
  background: rgba(0, 0, 0, 0.1);
}

.dark-theme .top-bar-search-suggestion:hover {
  background: rgba(255, 255, 255, 0.1);
}

.top-bar-search-suggestion i {
  margin-right: 0.5rem;
}

.top-bar-logo-img {
  width: 32px;
  height: 32px;
}

.top-bar-logo-td .top-bar-button {
  width: 72px;
}

.top-bar-menu-btn {
  visibility: hidden;
}

.sidebar-hidden .top-bar-menu-btn {
  visibility: visible;
}
</style>

