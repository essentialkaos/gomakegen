#!/bin/bash

################################################################################

main() {
  local sum1 sum2

  mv Makefile Makefile2

  ./gomakegen --mod --strip .

  [[ $? -ne 0 ]] && exit 1

  sum1=$(checksum "Makefile")
  sum2=$(checksum "Makefile2")

  if [[ "" != "" ]] ; then
    echo "Base Makefile differs from generated Makefile"
    exit 1
  fi

  echo "Base Makefile is equals generated Makefile"

  exit 0
}

checksum() {
  sha256sum < "$1" | cut -f1 -d" "
}

################################################################################

main "$@"
