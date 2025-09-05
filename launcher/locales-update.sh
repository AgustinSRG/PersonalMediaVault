#!/bin/bash

goi18n extract

goi18n merge active.*.toml

echo 'Translate the translate.*.toml files and run locales-done.sh'
