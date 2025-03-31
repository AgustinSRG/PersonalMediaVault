<template>
    <i :class="getClass(displayLoader, icon, extraClass)"><slot></slot></i>
</template>

<script lang="ts">
import { defineComponent } from "vue";

const DEFAULT_DELAY = 333;

export default defineComponent({
    name: "LoadingIcon",
    props: {
        icon: String,
        extraClass: String,
        loading: Boolean,
        delay: Number,
    },
    setup: function () {
        return {
            displayTimeout: null,
        };
    },
    data: function () {
        return {
            displayLoader: false,
        };
    },
    watch: {
        loading: function () {
            if (this.loading) {
                this.onLoadingTrue();
            } else {
                this.onLoadingFalse();
            }
        },
    },
    mounted: function () {
        if (this.loading) {
            this.onLoadingTrue();
        }
    },
    beforeUnmount: function () {
        if (this.displayTimeout) {
            clearTimeout(this.displayTimeout);
            this.displayTimeout = null;
        }
    },
    methods: {
        getClass: function (displayLoader: boolean, icon: string, extraClass: string): string {
            const classes = [];

            if (displayLoader) {
                classes.push("fa", "fa-spinner", "fa-spin");
            } else if (icon) {
                classes.push(icon);
            }

            if (extraClass) {
                classes.push(extraClass);
            }

            return classes.join(" ");
        },

        onLoadingTrue: function () {
            if (this.displayTimeout) {
                clearTimeout(this.displayTimeout);
                this.displayTimeout = null;
            }

            this.displayTimeout = setTimeout(() => {
                this.displayTimeout = null;
                this.displayLoader = true;
            }, DEFAULT_DELAY);
        },

        onLoadingFalse: function () {
            if (this.displayTimeout) {
                clearTimeout(this.displayTimeout);
                this.displayTimeout = null;
            }

            this.displayLoader = false;
        },
    },
});
</script>
