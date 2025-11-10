#/bin/bash

find -name \*.slint | xargs slint-tr-extractor --join-existing -d "pmv-gui" --package-name "pmv-gui" -o translations/es/LC_MESSAGES/pmv-gui.po
