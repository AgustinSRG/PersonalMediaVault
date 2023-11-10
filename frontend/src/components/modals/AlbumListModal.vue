<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus" :lock-close="busy">
        <div v-if="display" class="modal-dialog modal-sm" role="document">
            <div class="modal-header">
                <div class="modal-title">{{ $t("Albums") }}</div>
                <button type="button" class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>

            <div class="modal-body" v-if="loading"><i class="fa fa-spinner fa-spin"></i> {{ $t("Loading") }}...</div>

            <div class="modal-body with-menu" v-if="!loading">
                <div class="albums-modal-filter" v-if="!canWrite">
                    <input
                        type="text"
                        autocomplete="off"
                        @keydown="onFilterKeyDown"
                        @input="updateAlbums"
                        :disabled="busy"
                        v-model="filter"
                        class="form-control form-control-full-width auto-focus"
                        :placeholder="$t('Filter by name') + '...'"
                    />
                </div>
                <div class="albums-modal-filter with-edit-mode" v-if="canWrite">
                    <input
                        type="text"
                        autocomplete="off"
                        @keydown="onFilterKeyDown"
                        @input="updateAlbums"
                        :disabled="busy"
                        v-model="filter"
                        class="form-control auto-focus"
                        :placeholder="$t('Filter by name') + '...'"
                    />
                    <button
                        v-if="!editMode"
                        type="button"
                        :disabled="busy"
                        @click="changeEditMode"
                        class="album-edit-mode-btn"
                        :title="$t('Change to edit mode')"
                    >
                        <i class="fas fa-pencil-alt"></i>
                    </button>
                    <button
                        v-if="editMode"
                        type="button"
                        :disabled="busy"
                        @click="changeEditMode"
                        class="album-edit-mode-btn"
                        :title="$t('Change to view mode')"
                    >
                        <i class="fas fa-eye"></i>
                    </button>
                </div>
                <div class="albums-list-table-container">
                    <table class="modal-menu" v-if="editMode">
                        <tr v-if="albums.length === 0">
                            <td colspan="2" class="albums-menu-empty">
                                {{ $t("No albums found") }}
                            </td>
                        </tr>
                        <tr
                            v-for="a in albums"
                            :key="a.id"
                            class="modal-menu-item"
                            tabindex="0"
                            @click="clickOnAlbum(a)"
                            @keydown="clickOnEnter"
                        >
                            <td class="modal-menu-item-icon">
                                <i v-if="busy" class="fa fa-spinner fa-spin"></i>
                                <i v-else-if="a.added" class="far fa-square-check"></i>
                                <i v-else class="far fa-square"></i>
                            </td>
                            <td class="modal-menu-item-title">
                                {{ a.name }}
                            </td>
                        </tr>
                    </table>
                    <table class="modal-menu" v-if="!editMode">
                        <tr v-if="albums.length === 0">
                            <td colspan="2" class="albums-menu-empty">
                                {{ $t("No albums found") }}
                            </td>
                        </tr>
                        <tr
                            v-for="a in albums"
                            :key="a.id"
                            class="modal-menu-item"
                            tabindex="0"
                            @click="goToAlbum(a)"
                            @keydown="clickOnEnter"
                        >
                            <td class="modal-menu-item-icon">
                                <i class="fas fa-list-ol"></i>
                            </td>
                            <td class="modal-menu-item-title">
                                {{ a.name }}
                            </td>
                        </tr>
                    </table>
                </div>
            </div>

            <div class="modal-footer no-padding" v-if="!loading && editMode">
                <button type="button" @click="createAlbum" class="modal-footer-btn">
                    <i class="fas fa-plus"></i> {{ $t("Create album") }}
                </button>
            </div>
        </div>
        <AlbumCreateModal
            v-model:display="displayAlbumCreate"
            @new-album="onNewAlbum"
            @update:display="afterModalCreateClosed"
        ></AlbumCreateModal>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { AlbumsAPI } from "@/api/api-albums";
import { MediaAPI } from "@/api/api-media";
import { AlbumsController, EVENT_NAME_ALBUMS_LIST_UPDATE } from "@/control/albums";
import { AppEvents } from "@/control/app-events";
import { AppStatus } from "@/control/app-status";
import { AuthController, EVENT_NAME_UNAUTHORIZED } from "@/control/auth";
import { Request } from "@/utils/request";
import { Timeouts } from "@/utils/timeout";
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

import AlbumCreateModal from "../modals/AlbumCreateModal.vue";

export default defineComponent({
    components: {
        AlbumCreateModal,
    },
    name: "AlbumListModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    data: function () {
        return {
            albums: [],
            filter: "",

            mid: AppStatus.CurrentMedia,
            mediaAlbums: [],

            loading: true,
            busy: false,

            displayAlbumCreate: false,

            editMode: AuthController.CanWrite,
            canWrite: AuthController.CanWrite,
            editModeChanged: false,
        };
    },
    methods: {
        autoFocus: function () {
            nextTick(() => {
                const el = this.$el.querySelector(".auto-focus");

                if (el) {
                    el.focus();
                }
            });
        },

        afterModalCreateClosed: function (display: boolean) {
            if (!display && this.display) {
                this.autoFocus();
            }
        },

        load: function () {
            Timeouts.Abort("media-albums-load");
            Request.Abort("media-albums-load");

            if (!this.display) {
                return;
            }

            this.loading = true;

            if (AuthController.Locked) {
                return; // Vault is locked
            }

            Request.Pending("media-albums-load", MediaAPI.GetMediaAlbums(this.mid))
                .onSuccess((result) => {
                    this.mediaAlbums = result;
                    this.loading = false;
                    this.canWrite = AuthController.CanWrite;
                    if (!this.canWrite) {
                        this.editMode = false;
                    } else if (!this.editModeChanged) {
                        this.editMode = result.length === 0;
                    }
                    this.updateAlbums();
                    this.autoFocus();
                })
                .onRequestError((err) => {
                    Request.ErrorHandler()
                        .add(401, "*", () => {
                            AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                        })
                        .add(404, "*", () => {
                            this.$refs.modalContainer.close(true);
                        })
                        .add("*", "*", () => {
                            // Retry
                            Timeouts.Set("media-albums-load", 1500, this.load.bind(this));
                        })
                        .handle(err);
                })
                .onUnexpectedError((err) => {
                    console.error(err);
                    // Retry
                    Timeouts.Set("media-albums-load", 1500, this.load.bind(this));
                });
        },

        close: function () {
            this.$refs.modalContainer.close();
        },

        createAlbum: function () {
            this.displayAlbumCreate = true;
        },

        onNewAlbum: function (albumId, albumName) {
            this.filter = albumName;
            this.updateAlbums();
        },

        goToAlbum: function (album) {
            this.$refs.modalContainer.close(true);
            AppStatus.ClickOnAlbumByMedia(album.id, this.mid);
        },

        changeEditMode: function () {
            this.editMode = !this.editMode;
            this.editModeChanged = true;
            this.updateAlbums();
            this.autoFocus();
        },

        clickOnAlbum: function (album, backToText?: boolean) {
            if (this.busy) {
                return;
            }

            this.busy = true;

            if (album.added) {
                // Remove
                Request.Do(AlbumsAPI.RemoveMediaFromAlbum(album.id, this.mid))
                    .onSuccess(() => {
                        this.busy = false;
                        album.added = false;
                        AppEvents.ShowSnackBar(this.$t("Successfully removed from album"));
                        if (this.mediaAlbums.indexOf(album.id) >= 0) {
                            this.mediaAlbums.splice(this.mediaAlbums.indexOf(album.id), 1);
                        }
                        this.updateAlbums();
                        AlbumsController.OnChangedAlbum(album.id, true);
                        if (backToText && this.editMode) {
                            this.autoFocus();
                        }
                    })
                    .onRequestError((err) => {
                        this.busy = false;
                        Request.ErrorHandler()
                            .add(401, "*", () => {
                                AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                            })
                            .handle(err);
                    })
                    .onUnexpectedError((err) => {
                        this.busy = false;
                        console.error(err);
                    });
            } else {
                // Add
                Request.Do(AlbumsAPI.AddMediaToAlbum(album.id, this.mid))
                    .onSuccess(() => {
                        this.busy = false;
                        album.added = true;
                        AppEvents.ShowSnackBar(this.$t("Successfully added to album"));
                        if (this.mediaAlbums.indexOf(album.id) === -1) {
                            this.mediaAlbums.push(album.id);
                        }
                        this.updateAlbums();
                        AlbumsController.OnChangedAlbum(album.id, true);
                        if (backToText && this.editMode) {
                            this.changeEditMode();
                        }
                    })
                    .onRequestError((err) => {
                        this.busy = false;
                        Request.ErrorHandler()
                            .add(401, "*", () => {
                                AppEvents.Emit(EVENT_NAME_UNAUTHORIZED);
                            })
                            .add(400, "MAX_SIZE_REACHED", () => {
                                AppEvents.ShowSnackBar(
                                    this.$t("Error") +
                                        ":" +
                                        this.$t("The album reached the limit of 1024 elements. Please, consider creating another album."),
                                );
                            })
                            .add(403, "*", () => {
                                AppEvents.ShowSnackBar(this.$t("Error") + ":" + this.$t("Access denied"));
                            })
                            .handle(err);
                    })
                    .onUnexpectedError((err) => {
                        this.busy = false;
                        console.error(err);
                    });
            }
        },

        onUpdateStatus: function () {
            const changed = this.mid !== AppStatus.CurrentMedia;
            this.mid = AppStatus.CurrentMedia;
            if (changed) {
                this.updateAlbums();
            }
        },

        updateAlbums: function () {
            const mid = AppStatus.CurrentMedia;
            const filter = (this.filter + "").toLowerCase();
            this.albums = AlbumsController.GetAlbumsListMin()
                .filter((a) => {
                    return !filter || a.nameLowerCase.indexOf(filter) >= 0;
                })
                .map((a: any) => {
                    a.added = mid >= 0 && this.mediaAlbums.indexOf(a.id) >= 0;
                    return a;
                })
                .filter((a) => {
                    return this.editMode || a.added;
                })
                .sort((a, b) => {
                    if (filter) {
                        const aStarts = a.nameLowerCase.indexOf(filter) === 0;
                        const bStarts = b.nameLowerCase.indexOf(filter) === 0;

                        if (aStarts && !bStarts) {
                            return -1;
                        } else if (bStarts && !aStarts) {
                            return 1;
                        }
                    }
                    if (a.nameLowerCase < b.nameLowerCase) {
                        return -1;
                    } else if (a.nameLowerCase > b.nameLowerCase) {
                        return 1;
                    } else {
                        return 1;
                    }
                });
        },

        onFilterKeyDown: function (e: KeyboardEvent) {
            if (e.key === "Enter") {
                e.preventDefault();

                if (!this.filter) {
                    return;
                }

                if (this.albums.length === 0) {
                    return;
                }

                if (this.editMode) {
                    this.clickOnAlbum(this.albums[0], true);
                } else {
                    this.goToAlbum(this.albums[0]);
                }
            } else if (e.key === "Tab") {
                if (this.albums.length === 0) {
                    if (this.filter) {
                        e.preventDefault();
                    }
                    return;
                }

                if (this.filter === this.albums[0].name) {
                    return;
                }

                e.preventDefault();

                this.filter = this.albums[0].name;
            }
        },

        clickOnEnter: function (event) {
            if (event.key === "Enter") {
                event.preventDefault();
                event.stopPropagation();
                event.target.click();
            }
        },
    },
    mounted: function () {
        this._handles = Object.create(null);
        this._handles.albumsUpdateH = this.updateAlbums.bind(this);
        AppEvents.AddEventListener(EVENT_NAME_ALBUMS_LIST_UPDATE, this._handles.albumsUpdateH);

        this._handles.statusH = this.onUpdateStatus.bind(this);
        AppStatus.AddEventListener(this._handles.statusH);

        this.updateAlbums();
        this.load();

        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
            AlbumsController.Load();
        }
    },
    beforeUnmount: function () {
        AppEvents.RemoveEventListener(EVENT_NAME_ALBUMS_LIST_UPDATE, this._handles.albumsUpdateH);
        AppStatus.RemoveEventListener(this._handles.statusH);
        Timeouts.Abort("media-albums-load");
        Request.Abort("media-albums-load");
    },
    watch: {
        display: function () {
            this.displayAlbumCreate = false;
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
                AlbumsController.Load();
                this.load();
            } else {
                Timeouts.Abort("media-albums-load");
                Request.Abort("media-albums-load");
            }
        },
    },
});
</script>
