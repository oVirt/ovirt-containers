#!/bin/bash -ex

# The engine needs a 'max_connections' parameter larger than the default used
# by PostgreSQL, at most 150. As this parameter already has a value in the
# default 'postgresql.conf' file, we need to edit the file to fix it.
sed -i \
  -e "s/^\(max_connections\)\\s*=.*$/\1 = ${MAX_CONNECTIONS}/" \
  "${PGDATA}/postgresql.conf"

# Recent versions of the engine also need custom auto-vacuum parameters. The
# setup tool checks them, and fails the start of the engine if they aren't set.
# These are commented out by default, so we just need to append them to the
# file.
cat >> "${PGDATA}/postgresql.conf" <<.
autovacuum_vacuum_scale_factor = 0.01
autovacuum_analyze_scale_factor = 0.075
autovacuum_max_workers = 6
maintenance_work_mem = 65536
.
