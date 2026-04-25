#!/usr/bin/env bash
set -euo pipefail

hermes profile list
echo
hermes profile show default
echo
hermes profile show vuitton
echo
hermes profile show bottega
