<template>
    <div class="album-selector-container" tabindex="0" :class="{ expanded: expanded }" @keydown="onKeyDown">
        <div class="album-selector" :class="{ expanded: expanded, disabled: disabled }" @click="toggleExpand">
            <div class="album-selected-name">{{ getAlbumName(album, albumsListUpdateCount) }}</div>
            <div class="album-selector-chevron">
                <div class="chevron"></div>
            </div>
        </div>

        <div v-if="expanded" class="album-selector-suggestions-container">
            <div class="name-filter-input-container">
                <input
                    v-model="filter"
                    type="text"
                    class="form-control form-control-full-width name-filter-input"
                    autocomplete="off"
                    :placeholder="$t('Filter by name') + '...'"
                    @input="onFilterChanged"
                    @keydown="onInputKeyDown"
                />
            </div>
            <div class="album-selector-suggestions" @scroll.passive="onSuggestionsScroll">
                <a
                    v-for="s in suggestions"
                    :key="s.id"
                    target="_blank"
                    rel="noopener noreferrer"
                    :href="getSuggestionURL(s)"
                    class="album-selector-suggestion"
                    tabindex="0"
                    :title="s.name"
                    @click="clickSuggestion(s, $event)"
                    @keydown="onSuggestionKeyDown"
                >
                    <span>{{ s.name }}</span>
                </a>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";
import { generateURIQuery } from "@/utils/api";
import { BigListScroller } from "@/utils/big-list-scroller";
import { FocusTrap } from "@/utils/focus-trap";
import { filterToWords, matchSearchFilter, normalizeString } from "@/utils/normalize";
import { useVModel } from "@/utils/v-model";
import { defineComponent, nextTick } from "vue";

/**
 * Filtered album item
 */
interface AlbumItemFiltered {
    // ID
    id: number;

    // Name
    name: string;

    // Name (lowercase)
    nameLowerCase: string;

    // True if starts
    starts: boolean;

    // True if contains
    contains: boolean;
}

export default defineComponent({
    name: "AlbumSelect",
    props: {
        album: Number,
        disabled: Boolean,
    },
    emits: ["update:album"],
    setup(props) {
        return {
            albums: AlbumsController.GetAlbumsListMin(),
            albumState: useVModel(props, "album"),

            bigListScroller: null as BigListScroller<AlbumItemFiltered>,
            focusTrap: null as FocusTrap,

            blurTimeout: null as ReturnType<typeof setTimeout> | null,
            filterChangedTimeout: null as ReturnType<typeof setTimeout> | null,
        };
    },
    data: function () {
        return {
            suggestions: [] as AlbumItemFiltered[],
            filter: "",

            expanded: false,

            albumsListUpdateCount: 0,
        };
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_ALBUMS_LIST_UPDATE, this.onAlbumsListUpdated.bind(this));

        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this));

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
        if (this.focusTrap) {
            this.focusTrap.destroy();
        }

        if (this.blurTimeout) {
            clearTimeout(this.blurTimeout);
        }

        if (this.filterChangedTimeout) {
            clearTimeout(this.filterChangedTimeout);
        }
    },
    methods: {
        onAlbumsListUpdated: function () {
            this.albums = AlbumsController.GetAlbumsListMin();
            this.updateSuggestions();
            this.albumsListUpdateCount++;
        },

        getAlbumName: function (id: number, _updateCount: number): string {
            if (id < 0) {
                return "--";
            }

            const album = AlbumsController.AlbumsMap.get(id);

            if (!album) {
                return "";
            }

            return album.name;
        },

        onSuggestionsScroll: function (e: Event) {
            this.bigListScroller.checkElementScroll(e.target as HTMLElement);
        },

        getSuggestionURL: function (s: AlbumItemFiltered): string {
            return (
                window.location.protocol +
                "//" +
                window.location.host +
                window.location.pathname +
                generateURIQuery({
                    album: s.id + "",
                })
            );
        },

        clickSuggestion: function (s: AlbumItemFiltered, e: Event) {
            e.preventDefault();
            e.stopPropagation();

            this.albumState = s.id;
            this.closeInstantly();
            this.$el.focus();
        },

        onFilterChanged: function () {
            if (this.filterChangedTimeout) {
                return;
            }
            this.filterChangedTimeout = setTimeout(() => {
                this.filterChangedTimeout = null;
                this.updateSuggestions();
            }, 200);
        },

        onKeyDown: function (event: KeyboardEvent) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                this.toggleExpand();
            }
        },

        onInputKeyDown: function (event: KeyboardEvent) {
            if (event.key === "ArrowDown" && this.suggestions.length > 0) {
                event.preventDefault();
                const suggestionElement = this.$el.querySelector(".album-selector-suggestion");
                if (suggestionElement) {
                    suggestionElement.focus();
                }
            } else if (event.key === "Escape") {
                event.preventDefault();
                this.closeInstantly();
                this.$el.focus();
            }
            event.stopPropagation();
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
                    const searchInput = this.$el.querySelector(".name-filter-input");
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
                this.closeInstantly();
                this.$el.focus();
            }
        },

        updateSuggestions: function () {
            const albumFilter = normalizeString(this.filter).trim().toLowerCase();
            const albumFilterWords = filterToWords(albumFilter);

            const suggestions: AlbumItemFiltered[] = this.albums
                .map((a) => {
                    const i = albumFilter ? matchSearchFilter(a.name, albumFilter, albumFilterWords) : 0;
                    return {
                        id: a.id,
                        name: a.name,
                        nameLowerCase: a.nameLowerCase,
                        starts: i === 0,
                        contains: i >= 0,
                    };
                })
                .filter((a) => {
                    return a.starts || a.contains;
                })
                .sort((a, b) => {
                    if (a.starts && !b.starts) {
                        return -1;
                    } else if (b.starts && !a.starts) {
                        return 1;
                    } else if (a.nameLowerCase < b.nameLowerCase) {
                        return -1;
                    } else {
                        return 1;
                    }
                });

            if (!this.filter) {
                suggestions.unshift({
                    id: -1,
                    name: "--",
                    nameLowerCase: "--",
                    contains: true,
                    starts: true,
                });
            }

            this.bigListScroller.reset();
            this.bigListScroller.addElements(suggestions);

            nextTick(() => {
                const elem = this.$el.querySelector(".album-selector-suggestions");
                if (elem) {
                    elem.scrollTop = 0;
                }
            });
        },

        toggleExpand: function () {
            if (this.expanded) {
                this.closeInstantly();
                this.$el.focus();
            } else if (!this.disabled) {
                this.expand();
            }
        },

        expand: function () {
            if (this.blurTimeout) {
                clearTimeout(this.blurTimeout);
                this.blurTimeout = null;
            }
            this.expanded = true;
            this.updateSuggestions();
            if (this.focusTrap) {
                this.focusTrap.activate();
            }
            AlbumsController.Load();

            nextTick(() => {
                const elem = this.$el.querySelector(".name-filter-input");

                if (elem) {
                    elem.focus();
                    elem.select();
                }
            });
        },

        closeInstantly: function () {
            if (this.blurTimeout) {
                clearTimeout(this.blurTimeout);
                this.blurTimeout = null;
            }
            this.blurTimeout = null;
            this.expanded = false;
            this.filter = "";
            if (this.focusTrap) {
                this.focusTrap.deactivate();
            }
        },

        close: function () {
            if (this.blurTimeout) {
                clearTimeout(this.blurTimeout);
                this.blurTimeout = null;
            }
            this.blurTimeout = setTimeout(() => {
                this.blurTimeout = null;
                this.expanded = false;
                this.filter = "";
                if (this.focusTrap) {
                    this.focusTrap.deactivate();
                }
            }, 10);
        },
    },
});
</script>
