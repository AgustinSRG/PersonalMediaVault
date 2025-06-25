import eslint from "@eslint/js";
import eslintPluginVue from "eslint-plugin-vue";
import typescriptEslint from "typescript-eslint";
import globals from "globals";
import eslintConfigPrettier from "eslint-config-prettier";

export default typescriptEslint.config(
    { ignores: ["*.d.ts", "**/coverage", "**/dist"] },
    {
        extends: [eslint.configs.recommended, ...typescriptEslint.configs.recommended, ...eslintPluginVue.configs["flat/recommended"]],
        files: ["**/*.{ts,vue}"],
        languageOptions: {
            ecmaVersion: "latest",
            sourceType: "module",
            globals: globals.browser,
            parserOptions: {
                parser: typescriptEslint.parser,
            },
        },
        rules: {
            "no-empty": "off",
            "no-console": "off",
            "no-debugger": "warn",
            "no-useless-escape": "off",
            "@typescript-eslint/no-explicit-any": "off",
            "@typescript-eslint/no-empty-function": "off",
            "@typescript-eslint/no-unused-vars": [
                "error",
                { argsIgnorePattern: "^_", varsIgnorePattern: "^_", caughtErrorsIgnorePattern: "^_" },
            ],
            "@typescript-eslint/consistent-type-imports": ["error", { prefer: "type-imports" }],
            indent: "off",
            "vue/require-default-prop": "off",
            "vue/no-v-html": "off",
        },
    },
    eslintConfigPrettier,
);
