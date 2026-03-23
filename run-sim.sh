#!/bin/bash
# Login shell wrapper — runs the Sim game then disconnects
# Prefer shared binary if it exists, otherwise use built-in
if [ -x /opt/shared/sim ]; then
    exec /opt/shared/sim
else
    exec /opt/sim
fi
