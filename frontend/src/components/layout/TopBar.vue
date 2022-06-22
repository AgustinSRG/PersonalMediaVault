<template>
  <div class="top-bar">
    <div class="top-bar-logo-td">
      <button
        type="button"
        class="top-bar-button"
        :title="$t('Main menu')"
        @click="menu"
      >
        <i class="fas fa-bars"></i>
      </button>
      <span :title="$t('Personal Media Vault')" class="top-bar-title"
        ><i class="fab fa-youtube"></i> PMV</span
      >
      <span :title="$t('Personal Media Vault')" class="top-bar-title-min"
        ><i class="fab fa-youtube"></i> PMV</span
      >
    </div>
    <div class="top-bar-search-td">
      <div class="top-bar-center-div">
        <div class="top-bar-search-input-container">
          <input
            type="text"
            class="top-bar-search-input"
            spellcheck="false"
            autocorrect="off"
            autocomplete="off"
            autocapitalize="none"
            :placeholder="$t('Search by tag')"
            v-model="search"
            @change="goSearch"
          />
          <button
            type="button"
            class="top-bar-button top-bar-search-button"
            :title="$t('Search')"
            @click="goSearch"
          >
            <i class="fas fa-search"></i>
          </button>
        </div>
      </div>
    </div>
    <div class="top-bar-user-td">
      <button
        type="button"
        class="top-bar-button top-bar-botton-min-version"
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
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { defineComponent } from "vue";

export default defineComponent({
  name: "TopBar",
  emits: ["logout", "settings", "menu"],
  data: function () {
    return {
      search: AppStatus.CurrentSearch,
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
    },

    onSearchChanged: function () {
      this.search = AppStatus.CurrentSearch;
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
  background: #212121;

  display: flex;
  justify-content: space-between;
  align-items: center;

  white-space: nowrap;
}

.top-bar-logo-td {
  padding-left: 8px;
  text-align: left;
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
}

.top-bar-title-min {
  font-weight: bold;
  font-size: 24px;
  padding-left: 8px;
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
  color: rgba(255, 255, 255, 0.75);
  background: transparent;
  outline: none;
}

.top-bar-button:disabled {
  opacity: 0.7;
  cursor: default;
}

.top-bar-button:hover {
  color: white;
}

.top-bar-button:disabled:hover {
  color: rgba(255, 255, 255, 0.75);
}

.top-bar-button:focus {
  outline: none;
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
  border: 1px solid hsl(0, 0%, 18.82%);
  background: hsla(0, 0%, 100%, 0.08);
  display: flex;
  align-items: center;
  justify-content: center;
}

.top-bar-search-input {
  outline: none;
  width: calc(100% - 54px);
  color: white;
  background: hsl(0, 0%, 7%);
  height: 38px;
  border: none;
  font-size: 16px;
  margin: 0;
  padding: 1px 4px;
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

.top-bar-botton-min-version {
  display: none;
}

@media (max-width: 600px) {
  .top-bar-search-input-container {
    display: none;
  }

  .top-bar-botton-min-version {
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
</style>

