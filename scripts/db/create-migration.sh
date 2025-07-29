#!/bin/bash
# Create a new database migration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if migration name is provided
if [ -z "$1" ]; then
    echo -e "${RED}Error: Migration name is required${NC}"
    echo "Usage: ./create-migration.sh <migration_name>"
    echo "Example: ./create-migration.sh add_character_inventory"
    exit 1
fi

MIGRATION_NAME=$1
MIGRATIONS_DIR="../../mmorpg-backend/pkg/db/migrations"

# Get the next migration number
LAST_MIGRATION=$(ls -1 "$MIGRATIONS_DIR" 2>/dev/null | grep -E '^[0-9]{3}_.*\.up\.sql$' | sort -n | tail -1)
if [ -z "$LAST_MIGRATION" ]; then
    NEXT_NUMBER="001"
else
    LAST_NUMBER=$(echo "$LAST_MIGRATION" | grep -oE '^[0-9]{3}')
    NEXT_NUMBER=$(printf "%03d" $((10#$LAST_NUMBER + 1)))
fi

# Create file names
UP_FILE="${MIGRATIONS_DIR}/${NEXT_NUMBER}_${MIGRATION_NAME}.up.sql"
DOWN_FILE="${MIGRATIONS_DIR}/${NEXT_NUMBER}_${MIGRATION_NAME}.down.sql"

# Create up migration
cat > "$UP_FILE" << EOF
-- Migration: ${MIGRATION_NAME}
-- Created: $(date +%Y-%m-%d)
-- Phase: TODO - specify phase

BEGIN;

-- TODO: Add your forward migration SQL here

COMMIT;
EOF

# Create down migration
cat > "$DOWN_FILE" << EOF
-- Rollback: ${MIGRATION_NAME}
-- Created: $(date +%Y-%m-%d)

BEGIN;

-- TODO: Add your rollback SQL here

COMMIT;
EOF

echo -e "${GREEN}✓ Created migration files:${NC}"
echo -e "  ${YELLOW}${UP_FILE}${NC}"
echo -e "  ${YELLOW}${DOWN_FILE}${NC}"
echo ""
echo "Next steps:"
echo "1. Edit the migration files with your SQL"
echo "2. Test the migration locally"
echo "3. Add appropriate indexes and constraints"
echo "4. Ensure the rollback is complete and safe"