# Frontend customization

This document specifies the vault frontend customization capabilities implemented by the PersonalMediaVault backend, such as custom title, custom style or custom icons.

## Custom title and frontend style

You can customize the vault frontend title and also setup custom CSS code if you need it. Simply go to `Settings` and then `Advanced Settings`.

The custom title will affect the page base title.

The custom CSS code can be used to do complex stuff, like changing colors or font sizes. Note that this is not a recommended feature, since the style is already optimized for both dark and light theme.

## Custom favicon

By default, the PersonalMediaVault backend will serve the official icon.

You can change it by adding a custom `favicon.ico` into your vault folder.

You may have to clear cache to see the change, since the browser caches the icon to prevent requesting it multiple times.

## Custom logos

Similar to the favicon, by default, the official logos will be served.

If you want to change them, you must create a folder inside your vault folder with the name `img/icons`.

You can copy the [official ones](../../frontend/public/img/icons/) and modify them.

| File                | Description                                                   |
| ------------------- | ------------------------------------------------------------- |
| `favicon.png`       | Favicon used to display for the top bar. (Default format)     |
| `favicon.svg`       | Favicon used to display for the top bar. (Alternative format) |
| `favicon-32x32.png` | Small favicon image (32x32 px)                                |
| `favicon-16x16.png` | Small favicon image (16x16 px)                                |
