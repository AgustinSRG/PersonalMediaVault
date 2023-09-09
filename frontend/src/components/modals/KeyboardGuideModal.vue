<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
        <div v-if="display" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title">
                    {{ $t("Keyboard shortcuts") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding table-responsive">
                <table class="table table-text-overflow keyboard-help-table">
                    <thead>
                        <tr>
                            <th class="text-left">{{ $t("Action") }}</th>
                            <th class="text-left">{{ $t("Shortcut") }}</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <td>{{ $t("Open the menu") }}</td>
                            <td>
                                <span class="key-shortcut">⌘ ({{ $t("Control") }}) + M</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Search") }}</td>
                            <td>
                                <span class="key-shortcut">⌘ ({{ $t("Control") }}) + F</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Open vault settings") }}</td>
                            <td>
                                <span class="key-shortcut">⌘ ({{ $t("Control") }}) + S</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Close vault") }}</td>
                            <td>
                                <span class="key-shortcut">⌘ ({{ $t("Control") }}) + Q</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Close current modal or menu") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Escape") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Play") }} / {{ $t("Pause") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Space") }}</span>
                                {{ $t("or") }}
                                <span class="key-shortcut">K</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Fast forward 5 seconds") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Right arrow") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>
                                {{ $t("Fast forward a single frame (only when paused)") }}
                            </td>
                            <td>
                                <span class="key-shortcut">,</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Rewind 5 seconds") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Left arrow") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Rewind a single frame (only when paused)") }}</td>
                            <td>
                                <span class="key-shortcut">.</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Skip to the next time slice") }}</td>
                            <td>
                                <span class="key-shortcut">J</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Go back to the beginning of the time slice") }}</td>
                            <td>
                                <span class="key-shortcut">B</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Go to the beginning of the video / audio") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Home") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Go to the end of the video / audio") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("End") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Increase volume by") }} 5%</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Up arrow") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Increase volume by") }} 1%</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + {{ $t("Up arrow") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Decrease volume by") }} 5%</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Down arrow") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Decrease volume by") }} 1%</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + {{ $t("Down arrow") }}</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Mute") }} / {{ $t("Unmute") }}</td>
                            <td>
                                <span class="key-shortcut">M</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Toggle loop") }}</td>
                            <td>
                                <span class="key-shortcut">L</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Toggle slice loop") }}</td>
                            <td>
                                <span class="key-shortcut">X</span>
                            </td>
                        </tr>
                        <tr>
                            <td>{{ $t("Toggle full screen") }}</td>
                            <td>
                                <span class="key-shortcut">F</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Increment image scale by") }} 10%</td>
                            <td>
                                <span class="key-shortcut">⌘ ({{ $t("Control") }}) {{ $t("and") }} +</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Increment image scale by") }} 1%</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} {{ $t("and") }} +</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Decrement image scale by") }} 10%</td>
                            <td>
                                <span class="key-shortcut">⌘ ({{ $t("Control") }}) {{ $t("and") }} -</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Decrement image scale by") }} 1%</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} {{ $t("and") }} -</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Fit image into viewport size") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Space") }}</span>
                                {{ $t("or") }}
                                <span class="key-shortcut">K</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Hide or show image player controls") }}</td>
                            <td>
                                <span class="key-shortcut">C</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Scroll image") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Up arrow") }}</span>
                                {{ $t("or") }}
                                <span class="key-shortcut">{{ $t("Down arrow") }}</span>
                                {{ $t("or") }}
                                <span class="key-shortcut">{{ $t("Left arrow") }}</span>
                                {{ $t("or") }}
                                <span class="key-shortcut">{{ $t("Right arrow") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Scroll image to the top") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Home") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Scroll image to the bottom") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("End") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Scroll image to the left") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + {{ $t("Home") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Scroll image to the right") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + {{ $t("End") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Show player configuration") }}</td>
                            <td>
                                <span class="key-shortcut">S</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Open media information") }}</td>
                            <td>
                                <span class="key-shortcut">E</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Add current media to an album") }}</td>
                            <td>
                                <span class="key-shortcut">A</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Show media tags") }}</td>
                            <td>
                                <span class="key-shortcut">T</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Show media extended description") }}</td>
                            <td>
                                <span class="key-shortcut">I</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Go to next") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + {{ $t("Right arrow") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Go to previous") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + {{ $t("Left arrow") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Toggle album loop") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + L</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Toggle album shuffle") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + R</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Add or remove album to favorites") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Shift") }} + F</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Create new album (while in album list)") }}</td>
                            <td>
                                <span class="key-shortcut">+</span>
                            </td>
                        </tr>

                        <tr>
                            <td>
                                {{ $t("Close page at the right (in the split view mode)") }}
                            </td>
                            <td>
                                <span class="key-shortcut">Q</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Close player (in the split view mode)") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Backspace") }}</span>
                            </td>
                        </tr>

                        <tr>
                            <td>{{ $t("Delete media") }}</td>
                            <td>
                                <span class="key-shortcut">{{ $t("Delete") }}</span>
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
    name: "KeyboardGuideModal",
    emits: ["update:display"],
    props: {
        display: Boolean,
    },
    setup(props) {
        return {
            displayStatus: useVModel(props, "display"),
        };
    },
    methods: {
        close: function () {
            this.$refs.modalContainer.close();
        },
    },
    mounted: function () {
        if (this.display) {
            nextTick(() => {
                this.$el.focus();
            });
        }
    },
    watch: {
        display: function () {
            if (this.display) {
                nextTick(() => {
                    this.$el.focus();
                });
            }
        },
    },
});
</script>
