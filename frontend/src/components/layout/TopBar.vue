<template>
    <div class="top-bar" :class="{ 'search-focused': searchFocus }" tabindex="-1">
        <div class="top-bar-logo-td">
            <button type="button" class="top-bar-button top-bar-menu-btn" :title="$t('Main menu')" @click="menu">
                <i class="fas fa-bars"></i>
            </button>
            <img class="top-bar-logo-img" src="/img/icons/favicon.png" :alt="getAppLogoText(customLogo)" />
            <span :title="getAppTitle(customTitle)" class="top-bar-title">{{ getAppLogoText(customLogo) }}</span>
        </div>
        <div class="top-bar-search-td">
            <div class="top-bar-center-div">
                <form
                    class="top-bar-search-input-container"
                    :class="{ focused: searchFocus, 'has-search': hasSearch }"
                    tabindex="-1"
                    @submit="submitSearch"
                >
                    <input
                        v-model="search"
                        type="text"
                        class="top-bar-search-input"
                        name="pmv-search-input"
                        spellcheck="false"
                        autocorrect="off"
                        autocomplete="off"
                        autocapitalize="none"
                        :placeholder="$t('Search')"
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
                        <a
                            v-for="s in suggestions"
                            :key="s.key"
                            target="_blank"
                            rel="noopener noreferrer"
                            :href="getSuggestionUrl(s)"
                            class="top-bar-search-suggestion"
                            tabindex="0"
                            :title="s.name"
                            @click="clickSearch(s, $event)"
                            @keydown="onSuggestionKeyDown"
                        >
                            <i v-if="s.type === 'tag'" class="fas fa-tag"></i>
                            <i v-else-if="s.type === 'album'" class="fas fa-list-ol"></i>
                            <i v-else-if="s.type === 'search'" class="fas fa-search"></i>
                            <i v-else-if="s.type === 'media' && s.mediaType === 1" class="fas fa-image"></i>
                            <i v-else-if="s.type === 'media' && s.mediaType === 2" class="fas fa-video"></i>
                            <i v-else-if="s.type === 'media' && s.mediaType === 3" class="fas fa-headphones"></i>
                            <i v-else class="fas fa-ban"></i>
                            <span>{{ s.name }}</span>
                        </a>
                    </div>
                </form>
            </div>
        </div>
        <div class="top-bar-user-td">
            <button type="button" class="top-bar-button top-bar-button-small-version" :title="$t('Search')" @click="openSearch">
                <i class="fas fa-search"></i>
            </button>

            <button type="button" class="top-bar-button top-bar-button-dropdown" :title="$t('Help')" @click="help">
                <i class="fas fa-question"></i>
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
import { AppStatus } from "@/control/app-status";
import { AuthController } from "@/control/auth";
import { TagsController } from "@/control/tags";
import { defineComponent } from "vue";
import { FocusTrap } from "../../utils/focus-trap";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { BigListScroller } from "@/utils/big-list-scroller";
import { nextTick } from "vue";
import { getFrontendUrl } from "@/utils/api";
import { EVENT_NAME_APP_STATUS_CHANGED, EVENT_NAME_AUTH_CHANGED } from "@/control/app-events";

interface SearchBarSuggestion {
    key: string;

    id: string;

    name: string;
    nameLower: string;

    type: "tag" | "album" | "media" | "search";
    mediaType?: number;

    starts: boolean;
    contains: boolean;
}

export default defineComponent({
    name: "TopBar",
    emits: ["logout", "vault-settings", "account-settings", "menu", "menu-focus", "help"],
    setup() {
        return {
            blurTimeout: null as ReturnType<typeof setTimeout> | null,
            focusTrap: null as FocusTrap,
            bigListScroller: null as BigListScroller<SearchBarSuggestion>,
            findTagTimeout: null as ReturnType<typeof setTimeout> | null,
        };
    },
    data: function () {
        return {
            search: AppStatus.CurrentSearch,
            hasSearch: !!AppStatus.CurrentSearch,
            searchFocus: false,
            suggestions: [] as SearchBarSuggestion[],

            customTitle: AuthController.Title,
            customLogo: AuthController.Logo,
        };
    },

    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_APP_STATUS_CHANGED, this.onSearchChanged.bind(this));

        this.$listenOnAppEvent(EVENT_NAME_AUTH_CHANGED, () => {
            this.customTitle = AuthController.Title;
            this.customLogo = AuthController.Logo;
        });

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
    methods: {
        getAppTitle: function (customTitle: string) {
            return customTitle || this.$t("Personal Media Vault");
        },

        getAppLogoText: function (customLogo: string) {
            return customLogo || "PMV";
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
            this.focusSearch();

            nextTick(() => {
                const searchInput = this.$el.querySelector(".top-bar-search-input");
                if (searchInput) {
                    searchInput.focus();
                }
            });
        },

        help: function () {
            this.$emit("help");
        },

        goSearch: function () {
            AppStatus.GoToSearch(this.search);
            this.blurSearchInputElement();
        },

        blurSearchInputElement: function () {
            const el = this.$el.querySelector(".top-bar-search-input") as HTMLElement;
            if (el) {
                el.blur();
            }
        },

        goFindMedia: function () {
            const search = this.search;
            AppStatus.GoToSearch("");
            AppStatus.GoFindMedia(search);
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
            AlbumsController.Refresh();
            TagsController.Refresh();
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

        clickSearch: function (s: SearchBarSuggestion, e: Event) {
            e.preventDefault();
            this.blurSearch();
            this.selectSearch(s);
        },

        onSuggestionKeyDown: function (event: KeyboardEvent) {
            const target = event.target as HTMLElement;
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                target.click();
            } else if (event.key === "ArrowUp") {
                event.preventDefault();
                event.stopPropagation();
                const sibling = target.previousElementSibling as HTMLElement;
                if (sibling && sibling.focus) {
                    sibling.focus();
                } else {
                    const searchInput = this.$el.querySelector(".top-bar-search-input");
                    if (searchInput) {
                        searchInput.focus();
                    }
                }
            } else if (event.key === "ArrowDown") {
                event.preventDefault();
                event.stopPropagation();
                const sibling = target.nextElementSibling as HTMLElement;
                if (sibling && sibling.focus) {
                    sibling.focus();
                }
            } else if (event.key === "Escape") {
                event.preventDefault();
                event.stopPropagation();
                const searchInput = this.$el.querySelector(".top-bar-search-input");
                if (searchInput) {
                    searchInput.focus();
                }
            }
        },

        selectSearch: function (s: SearchBarSuggestion) {
            if (s.type === "album") {
                this.search = "";
                AppStatus.ClickOnAlbum(Number(s.id));
                this.blurSearchInputElement();
            } else if (s.type === "media") {
                this.search = "";
                AppStatus.ClickOnMedia(Number(s.id), false);
                this.blurSearchInputElement();
            } else if (s.type === "tag") {
                this.search = s.name;
                this.goSearch();
            } else {
                this.goFindMedia();
            }

            this.blurSearchInstantly();
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
                this.goFindMedia();
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

            let suggestions: SearchBarSuggestion[] = [];

            suggestions = suggestions.concat(
                Array.from(TagsController.Tags.entries())
                    .map((a): SearchBarSuggestion => {
                        const i = tagFilter ? normalizeString(a[1]).indexOf(tagFilter) : 0;
                        return {
                            key: "tag:" + a[0],
                            id: a[0] + "",
                            name: a[1],
                            nameLower: a[1],
                            starts: i === 0,
                            contains: i >= 0,
                            type: "tag",
                        };
                    })
                    .filter((a) => {
                        return a.starts || a.contains;
                    }),
            );

            suggestions = suggestions.concat(
                AlbumsController.GetAlbumsList()
                    .map((a): SearchBarSuggestion => {
                        const i = albumFilter ? matchSearchFilter(a.name, albumFilter, albumFilterWords) : 0;
                        return {
                            key: "album:" + a.id,
                            id: a.id + "",
                            name: a.name,
                            nameLower: a.name.toLowerCase(),
                            starts: i === 0,
                            contains: i >= 0,
                            type: "album",
                        };
                    })
                    .filter((a) => {
                        return a.starts || a.contains;
                    }),
            );

            if (AlbumsController.CurrentAlbumData) {
                suggestions = suggestions.concat(
                    AlbumsController.CurrentAlbumData.list
                        .map((a): SearchBarSuggestion => {
                            const i = albumFilter ? matchSearchFilter(a.title, albumFilter, albumFilterWords) : 0;
                            return {
                                key: "media:" + a.id,
                                id: a.id + "",
                                name: a.title,
                                nameLower: a.title.toLowerCase(),
                                starts: i === 0,
                                contains: i >= 0,
                                type: "media",
                                mediaType: a.type,
                            };
                        })
                        .filter((a) => {
                            return a.starts || a.contains;
                        }),
                );
            }

            suggestions = suggestions.sort((a, b) => {
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

            if (this.search) {
                suggestions.push({
                    key: "search:text",
                    id: "search:text",
                    name: this.search,
                    nameLower: this.search.toLowerCase(),
                    starts: true,
                    contains: true,
                    type: "search",
                });
            }

            this.bigListScroller.reset();
            this.bigListScroller.addElements(suggestions);

            nextTick(() => {
                const elem = this.$el.querySelector(".top-bar-search-suggestions");
                if (elem) {
                    elem.scrollTop = 0;
                }
            });
        },

        onSuggestionsScroll: function (e: Event) {
            this.bigListScroller.checkElementScroll(e.target as HTMLElement);
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
                } else if (this.suggestions.length > 0 && this.search === this.suggestions[0].name) {
                    event.preventDefault();
                    const suggestionElement = this.$el.querySelector(".top-bar-search-suggestion");
                    if (suggestionElement) {
                        suggestionElement.focus();
                    }
                }
            } else if (event.key.toUpperCase() === "F" && event.ctrlKey) {
                this.blurSearch();
            } else if (event.key === "Escape") {
                event.preventDefault();
                const inputElem = event.target as HTMLElement;
                if (inputElem && inputElem.blur) {
                    inputElem.blur();
                }
                this.blurSearch();
            } else if (event.key === "ArrowDown" && this.suggestions.length > 0) {
                event.preventDefault();
                const suggestionElement = this.$el.querySelector(".top-bar-search-suggestion");
                if (suggestionElement) {
                    suggestionElement.focus();
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

        getSuggestionUrl: function (s: SearchBarSuggestion): string {
            switch (s.type) {
                case "media":
                    return getFrontendUrl({
                        media: s.id,
                    });
                case "album":
                    return getFrontendUrl({
                        album: s.id,
                    });
                case "tag":
                    return getFrontendUrl({
                        page: "media",
                        search: s.name,
                    });
                default:
                    return getFrontendUrl({ page: "home" });
            }
        },
    },
});
</script>
