module.exports = {
    transpileDependencies: true,
    publicPath: "./",

    chainWebpack: (config) => {
        config.module.rule("images").set("parser", {
            dataUrlCondition: {
                maxSize: 1, // Disable
            },
        });

        config.module.rule("fonts").set("parser", {
            dataUrlCondition: {
                maxSize: 1, // Disable
            },
        });
    },
};
