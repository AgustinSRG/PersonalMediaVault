#/bin/bash

find -name \*.slint | xargs slint-tr-extractor --join-existing -d "pmv-launcher" --package-name "pmv-launcher" -o translations/es/LC_MESSAGES/pmv-gui.po
