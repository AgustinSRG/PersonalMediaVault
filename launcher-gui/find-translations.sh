#/bin/bash

PKG_NAME="pmv-gui"
LOCALES_DIR="translations"

update_translation_file() {
    local locale=$1
    sed -i "s/^#:.*/#:/" translations/$locale/LC_MESSAGES/$PKG_NAME.po
    find -name \*.slint | xargs slint-tr-extractor --join-existing -d "$PKG_NAME" --package-name "$PKG_NAME" -o translations/es/LC_MESSAGES/$PKG_NAME.po
}

# Find locales
find "$LOCALES_DIR" -maxdepth 1 -type d -print0 | while IFS= read -r -d $'\0' folder; do
    # Skip the parent directory itself
    if [ "$folder" = "$LOCALES_DIR" ]; then
        continue
    fi

    locale_to_process=`echo $folder | cut -d / -f 2`

    echo "Processing locale: $locale_to_process"

    update_translation_file $locale_to_process
done



