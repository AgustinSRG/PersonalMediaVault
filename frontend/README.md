# Personal Media Vault (Frontend)

This project contains the web interface for Personal Media Vault. The interface is inspired by YouTube, but adapted for the personal media use case.

Built using the [Vue.js](https://vuejs.org/) framework and [Vite](https://vitejs.dev/) asd the build tool.

## Compilation

First, install the dependencies using npm:

```sh
npm install
```

After installing the dependencies, compile the code for production with the following command:

```sh
npm run build
```

After compiling, the result will be saved in the `dist` folder.

## Development and testing

In order to run a development server to test the frontend, use:

```sh
npm run serve
```

In development, you need the backend running on another port. Make sure to set the `VITE_DEV_TEST_HOST` environment variable in the `.env.local` file so the frontend knows how to connect to the backend in development mode:

```conf
VITE_DEV_TEST_HOST=http://127.0.0.1:8000
```

In order to lint the code and test the typescript sources, type:

```sh
npm test
```

## Linter and prettier

This project uses [Eslint](https://eslint.org/) as the linter and [Prettier](https://prettier.io/) to automatically stylize the code. In order to run both, type the following command:

```sh
npm run lint
```

## Internationalization

Text translations can be found in the [src/locales](./src/locales/) folder as `json` files with a name structure like `locale-{LOCALE}.json`.

When using the `i18n` plugin, run the following script to detect its usages and update the translation files:

```sh
npm run update-translations
```

If there are missing translations, several `.txt` files will be created. You can use them or directly update the `.json` files.

After any update, remember to run the `update-translations` script.

## Font awesome

This project uses a script to create a font awesome subset.

When using new icons, make sure to run the script to update the font files:

```sh
npm run update-fa-subset
```
