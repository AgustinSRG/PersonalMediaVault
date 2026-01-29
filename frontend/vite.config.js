import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

import { VitePWA } from "vite-plugin-pwa";

// https://vitejs.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        VitePWA({
            injectRegister: null,
            manifest: {
                name: "Personal Media Vault",
                short_name: "PMV",
                theme_color: "#FF0000",
                background_color: "#000000",
                display: "standalone",
                start_url: "./",
                icons: [
                    {
                        src: "./img/icons/android-chrome-192x192.png",
                        sizes: "192x192",
                        type: "image/png",
                    },
                    {
                        src: "./img/icons/android-chrome-512x512.png",
                        sizes: "512x512",
                        type: "image/png",
                    },
                    {
                        src: "./img/icons/android-chrome-maskable-192x192.png",
                        sizes: "192x192",
                        type: "image/png",
                        purpose: "maskable",
                    },
                    {
                        src: "./img/icons/android-chrome-maskable-512x512.png",
                        sizes: "512x512",
                        type: "image/png",
                        purpose: "maskable",
                    },
                ],
            },
            workbox: {
                globPatterns: ["**/!(locale-*).{js,css,html,woff2}"],
                navigateFallbackAllowlist: [/^\/$/],
            },
        }),
    ],
    server: {
        port: 8080,
    },
    build: {
        assetsInlineLimit: 0,
        sourcemap: true,
        rollupOptions: {
            output: {
                chunkFileNames: (assetInfo) => {
                    if (assetInfo.name?.endsWith(".vue_vue_type_style_index_0_lang")) {
                        return `assets/${assetInfo.name.slice(0, -32)}-[hash:8].js`;
                    } else if (assetInfo.name?.endsWith(".vue_vue_type_script_setup_true_lang")) {
                        return `assets/${assetInfo.name.slice(0, -36)}-[hash:8].js`;
                    } else {
                        return "assets/[name]-[hash:8].js";
                    }
                },
            },
        },
    },
    resolve: {
        alias: {
            "@": fileURLToPath(new URL("./src", import.meta.url)),
        },
    },
    base: "./",
});
