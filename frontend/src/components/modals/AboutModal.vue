<template>
    <ModalDialogContainer ref="modalContainer" v-model:display="displayStatus">
        <div v-if="display" class="modal-dialog modal-lg" role="document">
            <div class="modal-header">
                <div class="modal-title about-modal-title">
                    <img class="about-modal-logo" src="/img/icons/favicon.png" alt="PMV" />
                    {{ $t("Personal Media Vault") }}
                </div>
                <button class="modal-close-btn" :title="$t('Close')" @click="close">
                    <i class="fas fa-times"></i>
                </button>
            </div>
            <div class="modal-body no-padding table-responsive">
                <table class="table table-text-overflow">
                    <tr>
                        <td>{{ $t("Version") }}</td>
                        <td>{{ version }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t("Version date") }}</td>
                        <td>{{ versionDate }}</td>
                    </tr>

                    <tr>
                        <td>{{ $t("Home page") }}</td>
                        <td>
                            <a :href="homePage" target="_blank" rel="noopener noreferrer">{{ homePage }}</a>
                        </td>
                    </tr>

                    <tr>
                        <td>{{ $t("Git repository") }}</td>
                        <td>
                            <a :href="gitRepo" target="_blank" rel="noopener noreferrer">{{ gitRepo }}</a>
                        </td>
                    </tr>

                    <tr>
                        <td>{{ $t("License") }}</td>
                        <td>
                            <a :href="license" target="_blank" rel="noopener noreferrer">{{ license }}</a>
                        </td>
                    </tr>
                </table>
            </div>
        </div>
    </ModalDialogContainer>
</template>

<script lang="ts">
import { defineComponent, nextTick } from "vue";
import { useVModel } from "../../utils/v-model";

export default defineComponent({
    name: "AboutModal",
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
            version: process.env.VUE_APP_VERSION || "-",
            versionDate: process.env.VUE_APP_VERSION_DATE || "-",
            homePage: process.env.VUE_APP_HOME_URL || "#",
            gitRepo: process.env.VUE_APP_GIT_URL || "#",
            license: process.env.VUE_APP_LICENSE_URL || "#",
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
