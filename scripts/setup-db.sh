#!/bin/bash

# Spectra-Red Database Setup Script
# This script initializes SurrealDB with schema and seed data

set -e  # Exit on error
set -u  # Exit on undefined variable

# Configuration
SURREALDB_URL="${SURREALDB_URL:-http://localhost:8000}"
SURREALDB_USER="${SURREALDB_USER:-root}"
SURREALDB_PASS="${SURREALDB_PASS:-root}"
SURREALDB_NS="${SURREALDB_NS:-spectra}"
SURREALDB_DB="${SURREALDB_DB:-intel}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Wait for SurrealDB to be ready
wait_for_surrealdb() {
    log_info "Waiting for SurrealDB to be ready at ${SURREALDB_URL}..."

    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        if curl -sf "${SURREALDB_URL}/health" > /dev/null 2>&1; then
            log_info "SurrealDB is ready!"
            return 0
        fi

        log_warn "Attempt ${attempt}/${max_attempts}: SurrealDB not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done

    log_error "SurrealDB failed to become ready after ${max_attempts} attempts"
    return 1
}

# Apply schema
apply_schema() {
    log_info "Applying database schema..."

    local schema_file="${1:-internal/db/schema/schema.surql}"

    if [ ! -f "$schema_file" ]; then
        log_warn "Schema file not found: ${schema_file}"
        log_warn "Skipping schema application (will be created in M1-T3)"
        return 0
    fi

    # Apply schema using surreal CLI or HTTP API
    if command -v surreal &> /dev/null; then
        log_info "Using surreal CLI to apply schema..."
        surreal import \
            --conn "${SURREALDB_URL}" \
            --user "${SURREALDB_USER}" \
            --pass "${SURREALDB_PASS}" \
            --ns "${SURREALDB_NS}" \
            --db "${SURREALDB_DB}" \
            "${schema_file}"
    else
        log_warn "surreal CLI not found, using HTTP API..."
        curl -X POST "${SURREALDB_URL}/sql" \
            -H "Accept: application/json" \
            -H "NS: ${SURREALDB_NS}" \
            -H "DB: ${SURREALDB_DB}" \
            -u "${SURREALDB_USER}:${SURREALDB_PASS}" \
            --data-binary "@${schema_file}"
    fi

    log_info "Schema applied successfully!"
}

# Load seed data
load_seed_data() {
    log_info "Loading seed data..."

    local seed_file="${1:-internal/db/schema/seed.surql}"

    if [ ! -f "$seed_file" ]; then
        log_warn "Seed file not found: ${seed_file}"
        log_warn "Skipping seed data (will be created in M1-T3)"
        return 0
    fi

    # Load seed data
    if command -v surreal &> /dev/null; then
        log_info "Using surreal CLI to load seed data..."
        surreal import \
            --conn "${SURREALDB_URL}" \
            --user "${SURREALDB_USER}" \
            --pass "${SURREALDB_PASS}" \
            --ns "${SURREALDB_NS}" \
            --db "${SURREALDB_DB}" \
            "${seed_file}"
    else
        log_warn "surreal CLI not found, using HTTP API..."
        curl -X POST "${SURREALDB_URL}/sql" \
            -H "Accept: application/json" \
            -H "NS: ${SURREALDB_NS}" \
            -H "DB: ${SURREALDB_DB}" \
            -u "${SURREALDB_USER}:${SURREALDB_PASS}" \
            --data-binary "@${seed_file}"
    fi

    log_info "Seed data loaded successfully!"
}

# Verify database setup
verify_setup() {
    log_info "Verifying database setup..."

    # Check database info
    local response=$(curl -s -X POST "${SURREALDB_URL}/sql" \
        -H "Accept: application/json" \
        -H "NS: ${SURREALDB_NS}" \
        -H "DB: ${SURREALDB_DB}" \
        -u "${SURREALDB_USER}:${SURREALDB_PASS}" \
        -d "INFO FOR DB;")

    if echo "$response" | grep -q "error"; then
        log_error "Database verification failed!"
        echo "$response"
        return 1
    fi

    log_info "Database info retrieved successfully"

    # Count records in key tables
    log_info "Counting records in tables..."

    local count_query="
        SELECT count() FROM common_port GROUP ALL;
        SELECT count() FROM country GROUP ALL;
        SELECT count() FROM asn GROUP ALL;
        SELECT count() FROM cloud_region GROUP ALL;
    "

    local counts=$(curl -s -X POST "${SURREALDB_URL}/sql" \
        -H "Accept: application/json" \
        -H "NS: ${SURREALDB_NS}" \
        -H "DB: ${SURREALDB_DB}" \
        -u "${SURREALDB_USER}:${SURREALDB_PASS}" \
        -d "$count_query")

    if command -v jq &> /dev/null; then
        log_info "Record counts:"
        echo "$counts" | jq -r '.[] | select(.result) | .result[] | "  \(.count) records"' || true
    else
        log_warn "jq not found, skipping detailed record count display"
    fi

    # Test a sample query
    log_info "Testing sample query (common ports)..."
    local sample_query="SELECT * FROM common_port WHERE label = 'ssh' LIMIT 1;"

    local sample_result=$(curl -s -X POST "${SURREALDB_URL}/sql" \
        -H "Accept: application/json" \
        -H "NS: ${SURREALDB_NS}" \
        -H "DB: ${SURREALDB_DB}" \
        -u "${SURREALDB_USER}:${SURREALDB_PASS}" \
        -d "$sample_query")

    if echo "$sample_result" | grep -q "ssh"; then
        log_info "Sample query successful - SSH port found"
    else
        log_warn "Sample query returned unexpected results"
    fi

    log_info "Database setup verified successfully!"
}

# Main execution
main() {
    log_info "Starting Spectra-Red database setup..."
    log_info "Target: ${SURREALDB_URL}"
    log_info "Namespace: ${SURREALDB_NS}"
    log_info "Database: ${SURREALDB_DB}"
    echo ""

    # Wait for SurrealDB
    wait_for_surrealdb || exit 1
    echo ""

    # Apply schema
    apply_schema "$@"
    echo ""

    # Load seed data
    load_seed_data "$@"
    echo ""

    # Verify setup
    verify_setup
    echo ""

    log_info "Database setup completed successfully!"
    log_info "You can access SurrealDB at: ${SURREALDB_URL}"
    log_info "Use credentials: ${SURREALDB_USER} / ${SURREALDB_PASS}"
}

# Run main function with all arguments
main "$@"
