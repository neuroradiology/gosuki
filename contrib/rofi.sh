#!/usr/bin/env sh

THEME_STR='window{ width: 100%; location: north; anchor: north; }'

selected=$(suki --format "<span lang=\"%u\">%T %t\t%u</span>\n" | rofi -theme-str "$THEME_STR" -markup-rows -i -p "bookmarks" -no-custom -dmenu)

exec xdg-open $(echo "$selected" | pup 'span attr{lang}')
