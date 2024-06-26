#!/bin/bash

################################################################################

main() {
  local sum1 sum2

  ./gomakegen --mod --strip . -o Makefile2

  [[ $? -ne 0 ]] && exit 1

  sum1=$(checksum "Makefile")
  sum2=$(checksum "Makefile2")

  if [[ "$sum1" != "$sum2" ]] ; then
    echo "Current Makefile differs from generated Makefile"
    diff Makefile Makefile2
    rm -f Makefile2
    exit 1
  fi

  echo "Current Makefile is equal to generated Makefile"

  rm -f Makefile2

  exit 0
}

checksum() {
  sha256sum < "$1" | cut -f1 -d" "
}

################################################################################

main "$@"
