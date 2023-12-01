module.exports = {
    root: true,
    env: {
        node: true,
        es2022: true,
    },
    extends: ["plugin:vue/vue3-essential", "eslint:recommended", "@vue/typescript/recommended"],
    parserOptions: {
        ecmaVersion: 2020,
    },
    rules: {
        "no-console": "off",
        "no-debugger": "warn",
        "no-useless-escape": "off",
        "@typescript-eslint/no-explicit-any": "off",
        "@typescript-eslint/no-empty-function": "off",
        "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_", varsIgnorePattern: "^_" }],
        indent: ["warn", 4, { SwitchCase: 1 }],
        "vue/max-len": [
            "warn",
            {
                code: 140,
                template: 140,
                tabWidth: 4,
                comments: 140,
                ignorePattern: "",
                ignoreComments: false,
                ignoreTrailingComments: false,
                ignoreUrls: false,
                ignoreStrings: false,
                ignoreTemplateLiterals: false,
                ignoreRegExpLiterals: false,
                ignoreHTMLAttributeValues: false,
                ignoreHTMLTextContents: false,
            },
        ],
    },
    overrides: [
        {
            files: ["**/__tests__/*.{j,t}s?(x)", "**/tests/unit/**/*.spec.{j,t}s?(x)"],
            env: {
                mocha: true,
            },
        },
    ],
};
