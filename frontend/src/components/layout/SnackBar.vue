<template>
    <div class="snackbar" :class="{ hidden: !shown }" @mouseenter="hide" @click="hide">{{ message }}</div>
</template>

<script lang="ts">
import { EVENT_NAME_SNACK_BAR } from "@/control/pages";
import { defineComponent } from "vue";

export default defineComponent({
    name: "SnackBar",
    setup() {
        return {
            timeout: null,
        };
    },
    data: function () {
        return {
            shown: false,
            message: "",
        };
    },
    methods: {
        show: function (msg: string) {
            if (this.timeout) {
                clearTimeout(this.timeout);
                this.timeout = null;
            }

            this.shown = true;
            this.message = msg;

            this.timeout = setTimeout(() => {
                this.shown = false;
            }, 3000);
        },

        hide: function () {
            if (this.timeout) {
                clearTimeout(this.timeout);
                this.timeout = null;
            }
            this.shown = false;
        },
    },
    mounted: function () {
        this.$listenOnAppEvent(EVENT_NAME_SNACK_BAR, this.show.bind(this));
    },
    beforeUnmount: function () {
        if (this.timeout) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
    },
});
</script>
