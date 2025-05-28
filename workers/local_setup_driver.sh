#!/bin/bash

DRIVER_PATH=$(python install_driver.py)
chmod +x "$DRIVER_PATH"
export WEB_DRIVER_PATH="$DRIVER_PATH"