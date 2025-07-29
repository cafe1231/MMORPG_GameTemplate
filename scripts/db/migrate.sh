#!/bin/bash
# Database migration helper script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="development"
COMMAND="up"
DATABASE_URL=""

# Function to show usage
show_usage() {
    echo "Usage: ./migrate.sh [OPTIONS] COMMAND"
    echo ""
    echo "Commands:"
    echo "  up              Run all pending migrations"
    echo "  down            Rollback the last migration"
    echo "  force VERSION   Force database to specific version"
    echo "  version         Show current migration version"
    echo "  create NAME     Create a new migration"
    echo ""
    echo "Options:"
    echo "  -e, --env       Environment (development|staging|production) [default: development]"
    echo "  -d, --database  Database URL (overrides environment default)"
    echo "  -h, --help      Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./migrate.sh up"
    echo "  ./migrate.sh -e production version"
    echo "  ./migrate.sh create add_character_stats"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--env)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -d|--database)
            DATABASE_URL="$2"
            shift 2
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        up|down|force|version|create)
            COMMAND="$1"
            shift
            if [[ "$COMMAND" == "force" || "$COMMAND" == "create" ]]; then
                if [[ -z "$1" ]]; then
                    echo -e "${RED}Error: $COMMAND requires an argument${NC}"
                    exit 1
                fi
                COMMAND_ARG="$1"
                shift
            fi
            ;;
        *)
            echo -e "${RED}Error: Unknown option $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

# Set database URL based on environment if not provided
if [[ -z "$DATABASE_URL" ]]; then
    case $ENVIRONMENT in
        development)
            DATABASE_URL="postgres://dev:dev@localhost:5432/mmorpg?sslmode=disable"
            ;;
        staging)
            DATABASE_URL="${STAGING_DATABASE_URL}"
            ;;
        production)
            DATABASE_URL="${PRODUCTION_DATABASE_URL}"
            ;;
        *)
            echo -e "${RED}Error: Unknown environment $ENVIRONMENT${NC}"
            exit 1
            ;;
    esac
fi

# Check if database URL is set
if [[ -z "$DATABASE_URL" ]]; then
    echo -e "${RED}Error: Database URL not set for environment $ENVIRONMENT${NC}"
    echo "Please set the appropriate environment variable or use -d option"
    exit 1
fi

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo -e "${RED}Error: migrate tool not found${NC}"
    echo "Install it with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

MIGRATIONS_PATH="../../mmorpg-backend/pkg/db/migrations"

# Execute command
case $COMMAND in
    up)
        echo -e "${BLUE}Running migrations...${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
        echo -e "${GREEN}✓ Migrations completed${NC}"
        ;;
    down)
        echo -e "${YELLOW}Rolling back last migration...${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down 1
        echo -e "${GREEN}✓ Rollback completed${NC}"
        ;;
    force)
        echo -e "${YELLOW}Warning: Forcing migration version to $COMMAND_ARG${NC}"
        read -p "Are you sure? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$COMMAND_ARG"
            echo -e "${GREEN}✓ Version forced to $COMMAND_ARG${NC}"
        else
            echo "Cancelled"
        fi
        ;;
    version)
        echo -e "${BLUE}Current migration version:${NC}"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
        ;;
    create)
        ./create-migration.sh "$COMMAND_ARG"
        ;;
    *)
        echo -e "${RED}Error: Unknown command $COMMAND${NC}"
        show_usage
        exit 1
        ;;
esac