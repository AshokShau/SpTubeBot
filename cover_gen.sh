#!/usr/bin/env bash

if [ -z "${1}" ]; then
  echo "Please supply a jpg image file"
  exit 1
fi

DESCRIPTION="Cover Artwork"
IMAGE_SOURCE="${1}"
IMAGE_MIME_TYPE="image/jpeg"
TARGET="${IMAGE_SOURCE%.j*}"
TYPE_ALBUM_COVER=3

print_hex() {
  local STRING="${1}"
  printf "0: %.8x" "${STRING}"
}

print_binary() {
  local STRING="${1}"
  echo -n "${STRING}" | xxd -r -g0
}

get_image_dimension() {
  local DIMENSION="${1}"
  file "${IMAGE_SOURCE}" | \
  awk -v dimension="${DIMENSION}" '
    BEGIN {
      dimensions["width"]=1;
      dimensions["height"]=2;
    }
    {
      for(dim_index=NF; dim_index>1; dim_index--) {
        if ($dim_index ~ /[0-9]{1,}x[0-9]{1,}/) {
          sub(/,/, "", $dim_index);
          split($dim_index, fields, "x");
          printf("%s", fields[dimensions[dimension]]);
          exit 0;
        }
      }
    }'
}

image_width() {
  echo -n $(get_image_dimension "width")
}

image_height() {
  echo -n $(get_image_dimension "height")
}

get_image_size() {
  local FILE=${@}
  echo -n "$(wc -c "${FILE}" | sed 's/^[ ]*//g' | cut -d ' ' -f1)"
}

add_to_target_binary() {
  local STRING="${1}"
  print_binary "$(print_hex ${STRING})" >> "${TARGET}.tmp"
}

add_to_target_direct() {
  local STRING="${1}"
  echo -n "${STRING}" >> "${TARGET}.tmp"
}

echo -n "" > "${TARGET}.tmp"

add_to_target_binary "${TYPE_ALBUM_COVER}"
add_to_target_binary "${#IMAGE_MIME_TYPE}"
add_to_target_direct "${IMAGE_MIME_TYPE}"
add_to_target_binary "${#DESCRIPTION}"
add_to_target_direct "${DESCRIPTION}"
add_to_target_binary 0
add_to_target_binary 0
add_to_target_binary 0
add_to_target_binary 0
add_to_target_binary "$(get_image_size ${IMAGE_SOURCE})"
cat "${IMAGE_SOURCE}" >> "${TARGET}.tmp"
if [ "$(uname)" == "Darwin" ]; then
    WRAP_OPT="--break=0"
else
    WRAP_OPT="--wrap=0"
fi
base64 "${WRAP_OPT}" "${TARGET}.tmp" > "${TARGET}.base64"
rm -f "${TARGET}.tmp"
