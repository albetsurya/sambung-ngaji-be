#!/bin/bash

BRANCH=$(git branch --show-current)

case "$BRANCH" in
  develop)
    cp .clasp.develop.json .clasp.json
    echo "Pushing to DEVELOPMENT..."
    clasp push
    ;;

  main)
    cp .clasp.main.json .clasp.json

    echo "WARNING: PUSHING TO PRODUCTION"
    read -p "Lanjutkan? (y/N): " CONFIRM

    if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
      echo "Push dibatalkan."
      exit 1
    fi

    clasp push
    ;;

  *)
    echo "Branch '$BRANCH' tidak dapat melakukan clasp push."
    exit 1
    ;;
esac