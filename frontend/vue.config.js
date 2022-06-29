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
    pwa: {
      name: "Personal Media Vault",
      themeColor: "#212121",
      manifestOptions: {
        display: "minimal-ui",
      },
    },
  }
})
