#!/bin/bash
# Login shell wrapper — runs the DnD game then disconnects
# Prefer shared binary if it exists, otherwise use built-in
if [ -x /opt/shared/dnd ]; then
    exec /opt/shared/dnd
else
    exec /opt/dnd
fi
