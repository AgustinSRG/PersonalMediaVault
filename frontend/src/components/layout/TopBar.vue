<template>
    <div class="top-bar" tabindex="-1">
        <div class="top-bar-logo-td">
            <button type="button" class="top-bar-button top-bar-menu-btn" :title="$t('Main menu')" @click="menu">
                <i class="fas fa-bars"></i>
            </button>
            <img class="top-bar-logo-img" src="/img/icons/favicon.png" alt="PMV" />
            <span :title="getAppTitle()" class="top-bar-title">PMV</span>
        </div>
        <div class="top-bar-search-td">
            <div class="top-bar-center-div">
                <form
                    class="top-bar-search-input-container"
                    :class="{ focused: searchFocus, 'has-search': hasSearch }"
                    @submit="submitSearch"
                    tabindex="-1"
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
                    />
                    <button
                        v-if="hasSearch"
                        type="button"
                        class="top-bar-search-clear-btn"
                        :title="$t('Clear search')"
                        @click="clearSearch"
                    >
                        <i class="fas fa-times"></i>
                    </button>
                    <button type="submit" class="top-bar-button top-bar-search-button" :title="$t('Search')" @focus="blurSearch">
                        <i class="fas fa-search"></i>
                    </button>
                    <div
                        class="top-bar-search-suggestions"
                        :class="{ hidden: suggestions.length === 0 }"
                        @scroll.passive="onSuggestionsScroll"
                    >
                        <div
                            v-for="s in suggestions"
                            :key="s.key"
                            class="top-bar-search-suggestion"
                            @click="clickSearch(s)"
                            :title="s.name"
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
                class="top-bar-button top-bar-button-dropdown top-bar-button-large-version"
                :title="$t('Help')"
                @click="help"
            >
                <i class="fas fa-question"></i>
            </button>

            <button type="button" class="top-bar-button top-bar-button-small-version" :title="$t('Search')" @click="openSearch">
                <i class="fas fa-search"></i>
            </button>

            <button type="button" class="top-bar-button top-bar-button-dropdown" :title="$t('Vault settings')" @click="vaultSettings">
                <i class="fas fa-cog"></i>
            </button>

            <button type="button" class="top-bar-button top-bar-button-dropdown" :title="$t('Account settings')" @click="accountSettings">
                <i class="fas fa-user-cog"></i>
            </button>
        </div>
    </div>
</template>

<script lang="ts">
import { AlbumsController } from "@/control/albums";
import { AppStatus, EVENT_NAME_APP_STATUS_CHANGED } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { TagsController } from "@/control/tags";
import { defineComponent } from "vue";
import { FocusTrap } from "../../utils/focus-trap";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { BigListScroller } from "@/utils/big-list-scroller";
import { nextTick } from "vue";
import { EVENT_NAME_SEARCH_MODAL_SUBMIT } from "@/control/pages";

export default defineComponent({
    name: "TopBar",
    emits: ["logout", "vault-settings", "account-settings", "menu", "menu-focus", "search-open", "help"],
    setup() {
        return {
            blurTimeout: null,
            focusTrap: null as FocusTrap,
            bigListScroller: null as BigListScroller,
            findTagTimeout: null,
        };
    },
    data: function () {
        return {
            search: AppStatus.CurrentSearch,
            hasSearch: !!AppStatus.CurrentSearch,
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

        vaultSettings: function () {
            this.$emit("vault-settings");
        },

        accountSettings: function () {
            this.$emit("account-settings");
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
            this.hasSearch = !!AppStatus.CurrentSearch;
        },

        focusSearch: function () {
            if (this.blurTimeout) {
                clearTimeout(this.blurTimeout);
                this.blurTimeout = null;
            }
            this.searchFocus = true;
            this.$el.querySelector(".top-bar-search-input").select();
            this.updateSuggestions();
            if (this.focusTrap) {
                this.focusTrap.activate();
            }
            AlbumsController.Load();
            TagsController.Load();
        },

        blurSearchInstantly: function () {
            if (this.blurTimeout) {
                clearTimeout(this.blurTimeout);
                this.blurTimeout = null;
            }
            this.blurTimeout = null;
            this.searchFocus = false;
            if (this.focusTrap) {
                this.focusTrap.deactivate();
            }
        },

        blurSearch: function () {
            if (this.blurTimeout) {
                clearTimeout(this.blurTimeout);
                this.blurTimeout = null;
            }
            this.blurTimeout = setTimeout(() => {
                this.blurTimeout = null;
                this.searchFocus = false;
                if (this.focusTrap) {
                    this.focusTrap.deactivate();
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

        submitSearch: function (event?: Event) {
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

        clearSearch: function () {
            this.blurSearchInstantly();
            this.search = "";
            this.goSearch();
        },

        updateSuggestions: function () {
            const tagFilter = normalizeString(
                this.search
                    .replace(/[\n\r]/g, " ")
                    .trim()
                    .replace(/[\s]/g, "_")
                    .toLowerCase(),
            );
            const albumFilter = normalizeString(this.search).trim().toLowerCase();
            const albumFilterWords = filterToWords(albumFilter);

            const suggestions = Array.from(TagsController.Tags.entries())
                .map((a: any) => {
                    const i = tagFilter ? normalizeString(a[1]).indexOf(tagFilter) : 0;
                    return {
                        key: "tag:" + a[0],
                        id: a[0],
                        name: a[1],
                        nameLower: a[1],
                        starts: i === 0,
                        contains: i >= 0,
                        type: "tag",
                    };
                })
                .concat(
                    AlbumsController.GetAlbumsList().map((a) => {
                        const i = albumFilter ? matchSearchFilter(a.name, albumFilter, albumFilterWords) : 0;
                        return {
                            key: "album:" + a.id,
                            id: a.id,
                            name: a.name,
                            nameLower: a.name.toLowerCase(),
                            starts: i === 0,
                            contains: i >= 0,
                            type: "album",
                        };
                    }),
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

            this.bigListScroller.reset();
            this.bigListScroller.addElements(suggestions);

            nextTick(() => {
                const elem = this.$el.querySelector(".top-bar-search-suggestions");
                if (elem) {
                    elem.scrollTop = 0;
                }
            });
        },

        onSuggestionsScroll: function (e) {
            this.bigListScroller.checkElementScroll(e.target);
        },

        onSearchInput: function () {
            if (this.findTagTimeout) {
                return;
            }
            this.findTagTimeout = setTimeout(() => {
                this.findTagTimeout = null;
                this.updateSuggestions();
            }, 200);
        },

        onKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Tab" && this.search && !event.shiftKey) {
                if (this.suggestions.length > 0 && this.search !== this.suggestions[0].name) {
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
                this.$emit("menu-focus");
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
                if (event.shiftKey) {
                    this.accountSettings();
                } else {
                    this.vaultSettings();
                }

                return true;
            }

            if (event.key.toUpperCase() === "H" && event.ctrlKey && event.shiftKey) {
                this.help();
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
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onSearchChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_SEARCH_MODAL_SUBMIT, this.onSearchModalSubmit.bind(this));

        this.$addKeyboardHandler(this.handleGlobalKey.bind(this));

        this.focusTrap = new FocusTrap(this.$el.querySelector(".top-bar-search-input-container"), this.blurSearch.bind(this));

        this.bigListScroller = new BigListScroller(BigListScroller.GetWindowSize(9), {
            get: (): any[] => {
                return this.suggestions;
            },
            set: (list) => {
                this.suggestions = list;
            },
        });
    },

    beforeUnmount: function () {
        if (this.findTagTimeout) {
            clearTimeout(this.findTagTimeout);
        }

        if (this.blurTimeout) {
            clearTimeout(this.blurTimeout);
            this.blurTimeout = null;
        }

        if (this.focusTrap) {
            this.focusTrap.destroy();
        }
    },
});
</script>
