<template>
    <div
        class="image-player-config"
        :class="{ hidden: !shown }"
        tabindex="-1"
        role="dialog"
        @click="stopPropagationEvent"
        @dblclick="stopPropagationEvent"
        @mousedown="stopPropagationEvent"
        @touchstart="stopPropagationEvent"
        @contextmenu="stopPropagationEvent"
        @mouseenter="enterConfig"
        @mouseleave="leaveConfig"
        @keydown="keyDownHandle"
    >
        <table v-if="page === ''">
            <tbody>
                <tr>
                    <td>
                        <i class="fas fa-eye-slash icon-config"></i>
                        <b>{{ $t("Hide image notes") }}</b>
                    </td>
                    <td class="td-right">
                        <ToggleSwitch v-model:val="hideNotes" @update:val="changeNotesVisible"></ToggleSwitch>
                    </td>
                </tr>
                <tr class="tr-button" tabindex="0" @click="goToResolutions" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-photo-film icon-config"></i>
                        <b>{{ $t("Quality") }}</b>
                    </td>
                    <td class="td-right">
                        {{ renderResolution(resolution, rTick) }}
                        <i class="fas fa-chevron-right arrow-config"></i>
                    </td>
                </tr>

                <tr class="tr-button" tabindex="0" @click="goToBackgrounds" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-palette icon-config"></i>
                        <b>{{ $t("Background") }}</b>
                    </td>
                    <td class="td-right">
                        {{ renderBackground(background) }}
                        <i class="fas fa-chevron-right arrow-config"></i>
                    </td>
                </tr>

                <tr class="tr-button" tabindex="0" @click="goToAutoNext" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-forward icon-config"></i>
                        <b>{{ $t("Auto next") }}</b>
                    </td>
                    <td class="td-right">
                        {{ renderAutoNext(autoNext) }}
                        <i class="fas fa-chevron-right arrow-config"></i>
                    </td>
                </tr>
            </tbody>
        </table>
        <table v-if="page === 'resolution'">
            <tbody>
                <tr class="tr-button" tabindex="0" @click="goBack" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Quality") }}</b>
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr class="tr-button" tabindex="0" @click="changeResolution(-1)" @keydown="clickOnEnter">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': -1 !== resolution }"></i>
                        {{ renderResolution(-1, rTick) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr
                    v-for="(r, i) in metadata.resolutions"
                    :key="i"
                    class="tr-button"
                    tabindex="0"
                    @keydown="clickOnEnter"
                    @click="changeResolution(i)"
                >
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': i !== resolution }"></i>
                        {{ renderResolution(i, rTick) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>
        <table v-if="page === 'background'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Background") }}</b>
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr v-for="b in bgOptions" :key="b" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeBackground(b)">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': b !== background }"></i>
                        {{ renderBackground(b) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>
        <table v-if="page === 'auto-next'">
            <tbody>
                <tr class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="goBack">
                    <td>
                        <i class="fas fa-chevron-left icon-config"></i>
                        <b>{{ $t("Auto next") }}</b>
                    </td>
                    <td class="td-right"></td>
                </tr>
                <tr v-for="b in autoNextOptions" :key="b" class="tr-button" tabindex="0" @keydown="clickOnEnter" @click="changeAutoNext(b)">
                    <td>
                        <i class="fas fa-check icon-config" :class="{ 'check-uncheck': b !== autoNext }"></i>
                        {{ renderAutoNext(b) }}
                    </td>
                    <td class="td-right"></td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<script lang="ts">
import { getAutoNextTime, getImageNotesVisible, setAutoNextTime, setImageNotesVisible } from "@/control/player-preferences";
import { defineComponent, nextTick, PropType } from "vue";
import { useVModel } from "../../utils/v-model";
import { FocusTrap } from "../../utils/focus-trap";
import ToggleSwitch from "../utils/ToggleSwitch.vue";
import { MediaData, MediaResolution } from "@/api/models";

export default defineComponent({
    name: "ImagePlayerConfig",
    components: { ToggleSwitch },
    props: {
        shown: Boolean,
        metadata: Object as PropType<MediaData>,
        resolution: Number,
        background: String,
        rTick: Number,
    },
    emits: ["update:shown", "update:resolution", "update:background", "update-auto-next", "update-notes-visible", "enter", "leave"],
    setup(props) {
        return {
            focusTrap: null as FocusTrap,
            shownState: useVModel(props, "shown"),
            resolutionState: useVModel(props, "resolution"),
            backgroundState: useVModel(props, "background"),
        };
    },
    data: function () {
        return {
            page: "",
            resolutions: [] as MediaResolution[],
            bgOptions: ["default", "black", "white"],
            autoNext: getAutoNextTime(),
            autoNextOptions: [0, 3, 5, 10, 15, 20, 25, 30],
            hideNotes: !getImageNotesVisible(),
        };
    },
    watch: {
        shown: function () {
            this.page = "";
            if (this.shown) {
                this.focusTrap.activate();
                nextTick(() => {
                    this.$el.focus();
                });
            } else {
                this.focusTrap.deactivate();
            }
        },
        rTick: function () {
            this.updateResolutions();
        },
    },
    mounted: function () {
        this.focusTrap = new FocusTrap(this.$el, this.close.bind(this), "player-settings-no-trap");
        this.updateResolutions();
    },
    beforeUnmount: function () {
        this.focusTrap.destroy();
    },
    methods: {
        changeResolution: function (i: number) {
            this.resolutionState = i;
        },

        changeBackground: function (b: string) {
            this.backgroundState = b;
        },

        changeAutoNext: function (b: number) {
            this.autoNext = b;
            setAutoNextTime(b);
            this.$emit("update-auto-next");
        },

        changeNotesVisible: function () {
            setImageNotesVisible(!this.hideNotes);
            this.$emit("update-notes-visible", !this.hideNotes);
        },

        enterConfig: function () {
            this.$emit("enter");
        },

        leaveConfig: function () {
            this.$emit("leave");
        },

        focus: function () {
            nextTick(() => {
                this.$el.focus();
            });
        },

        goBack: function () {
            this.page = "";
            this.focus();
        },

        goToResolutions: function () {
            this.page = "resolution";
            this.focus();
        },

        goToBackgrounds: function () {
            this.page = "background";
            this.focus();
        },

        goToAutoNext: function () {
            this.page = "auto-next";
            this.focus();
        },

        renderBackground: function (b: string) {
            switch (b) {
                case "white":
                    return this.$t("White");
                case "black":
                    return this.$t("Black");
                default:
                    return this.$t("Default (Theme)");
            }
        },

        renderAutoNext: function (s: number) {
            if (!isNaN(s) && isFinite(s) && s > 0) {
                if (s === 1) {
                    return s + " " + this.$t("second");
                } else {
                    return s + " " + this.$t("seconds");
                }
            } else {
                return this.$t("Disabled");
            }
        },

        renderResolution: function (res: number, rTick: number) {
            if (rTick < 0 || !this.metadata) {
                return this.$t("Unknown");
            }
            if (res < 0) {
                return (
                    this.metadata.width +
                    "x" +
                    this.metadata.height +
                    " (" +
                    this.$t("Original") +
                    ")" +
                    (this.metadata.encoded ? "" : " (" + this.$t("Pending") + ")")
                );
            } else {
                const resData = this.metadata.resolutions[res];

                let width = this.metadata.width;
                let height = this.metadata.height;

                if (width > height) {
                    const proportionalHeight = Math.round((height * resData.width) / width);

                    if (proportionalHeight > resData.height) {
                        width = Math.round((width * resData.height) / height);
                        height = resData.height;
                    } else {
                        width = resData.width;
                        height = proportionalHeight;
                    }
                } else {
                    const proportionalWidth = Math.round((width * resData.height) / height);

                    if (proportionalWidth > resData.width) {
                        height = Math.round((height * resData.width) / width);
                        width = resData.width;
                    } else {
                        width = proportionalWidth;
                        height = resData.height;
                    }
                }

                if (resData) {
                    return width + "x" + height + "" + (resData.ready ? "" : " (" + this.$t("Pending") + ")");
                } else {
                    return this.$t("Unknown");
                }
            }
        },
        updateResolutions: function () {
            if (this.metadata && this.metadata.resolutions) {
                this.resolutions = this.metadata.resolutions.slice();
            } else {
                this.resolutions = [];
            }
        },

        close: function () {
            this.shownState = false;
        },

        keyDownHandle: function (e: KeyboardEvent) {
            if (e.ctrlKey) {
                return;
            }
            if (e.key === "Escape") {
                this.close();
                e.stopPropagation();
            }
        },
    },
});
</script>
