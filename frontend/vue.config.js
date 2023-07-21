const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
    transpileDependencies: true,
    publicPath: './',

    pluginOptions: {
        i18n: {
            locale: 'en',
            fallbackLocale: 'en',
            localeDir: 'locales',
            enableLegacy: true,
            runtimeOnly: false,
            compositionOnly: true,
            fullInstall: true
        },
    },

    pwa: {
        name: "Personal Media Vault",
        themeColor: "#FF0000",
        manifestOptions: {
            display: "minimal-ui",
        },
    },

    chainWebpack: config => {
        config.module
            .rule('images')
            .set('parser', {
                dataUrlCondition: {
                    maxSize: 1 // Disable
                }
            });

        config.module
            .rule('fonts')
            .set('parser', {
                dataUrlCondition: {
                    maxSize: 1 // Disable
                }
            })
    },
})
