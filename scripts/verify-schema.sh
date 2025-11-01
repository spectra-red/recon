#!/bin/bash

# Spectra-Red Schema Verification Script
# This script validates that the schema and seed data are correctly loaded

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
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Execute SurrealQL query
query() {
    local sql="$1"
    curl -s -X POST "${SURREALDB_URL}/sql" \
        -H "Accept: application/json" \
        -H "NS: ${SURREALDB_NS}" \
        -H "DB: ${SURREALDB_DB}" \
        -u "${SURREALDB_USER}:${SURREALDB_PASS}" \
        -d "$sql"
}

# Test table existence
test_table_exists() {
    local table="$1"
    log_test "Checking if table '$table' exists..."

    local result=$(query "INFO FOR TABLE $table;")

    if echo "$result" | grep -q "error"; then
        log_fail "Table '$table' does not exist"
        return 1
    else
        log_pass "Table '$table' exists"
        return 0
    fi
}

# Test record count
test_record_count() {
    local table="$1"
    local expected_min="$2"
    log_test "Checking record count in table '$table' (expecting >= $expected_min)..."

    local result=$(query "SELECT count() FROM $table GROUP ALL;")

    if command -v jq &> /dev/null; then
        local count=$(echo "$result" | jq -r '.[0].result[0].count' 2>/dev/null || echo "0")

        if [ "$count" -ge "$expected_min" ]; then
            log_pass "Table '$table' has $count records (>= $expected_min)"
            return 0
        else
            log_fail "Table '$table' has only $count records (expected >= $expected_min)"
            return 1
        fi
    else
        log_warn "jq not installed, skipping count verification for $table"
        return 0
    fi
}

# Test index exists
test_index_exists() {
    local table="$1"
    local index="$2"
    log_test "Checking if index '$index' exists on table '$table'..."

    local result=$(query "INFO FOR TABLE $table;")

    if echo "$result" | grep -q "$index"; then
        log_pass "Index '$index' exists on table '$table'"
        return 0
    else
        log_fail "Index '$index' not found on table '$table'"
        return 1
    fi
}

# Test specific record exists
test_record_exists() {
    local table="$1"
    local field="$2"
    local value="$3"
    log_test "Checking if record exists in '$table' where $field = '$value'..."

    local result=$(query "SELECT * FROM $table WHERE $field = '$value' LIMIT 1;")

    if echo "$result" | grep -q "$value"; then
        log_pass "Record found in '$table' where $field = '$value'"
        return 0
    else
        log_fail "No record found in '$table' where $field = '$value'"
        return 1
    fi
}

# Main verification
main() {
    log_info "Starting Spectra-Red Schema Verification"
    log_info "Target: ${SURREALDB_URL}"
    log_info "Namespace: ${SURREALDB_NS}"
    log_info "Database: ${SURREALDB_DB}"
    echo ""

    # ========================================
    # Test Core Tables Existence
    # ========================================
    echo "=== Testing Core Tables ==="
    test_table_exists "host"
    test_table_exists "port"
    test_table_exists "service"
    test_table_exists "banner"
    test_table_exists "tls_cert"
    echo ""

    # ========================================
    # Test Vulnerability Tables
    # ========================================
    echo "=== Testing Vulnerability Tables ==="
    test_table_exists "vuln"
    test_table_exists "vuln_doc"
    echo ""

    # ========================================
    # Test Geography Tables
    # ========================================
    echo "=== Testing Geography Tables ==="
    test_table_exists "city"
    test_table_exists "region"
    test_table_exists "country"
    test_table_exists "asn"
    test_table_exists "cloud_region"
    test_table_exists "common_port"
    echo ""

    # ========================================
    # Test Relationship Tables
    # ========================================
    echo "=== Testing Relationship Tables ==="
    test_table_exists "HAS"
    test_table_exists "RUNS"
    test_table_exists "EVIDENCED_BY"
    test_table_exists "AFFECTED_BY"
    test_table_exists "IN_CITY"
    test_table_exists "IN_REGION"
    test_table_exists "IN_COUNTRY"
    test_table_exists "IN_ASN"
    test_table_exists "IN_CLOUD_REGION"
    test_table_exists "IS_COMMON"
    test_table_exists "OBSERVED_AT"
    echo ""

    # ========================================
    # Test Seed Data Record Counts
    # ========================================
    echo "=== Testing Seed Data Record Counts ==="
    test_record_count "common_port" 28
    test_record_count "country" 25
    test_record_count "asn" 20
    test_record_count "cloud_region" 26
    test_record_count "region" 14
    test_record_count "city" 25
    test_record_count "vuln" 5
    test_record_count "vuln_doc" 3
    echo ""

    # ========================================
    # Test Key Indices
    # ========================================
    echo "=== Testing Key Indices ==="
    test_index_exists "host" "idx_host_ip"
    test_index_exists "port" "idx_port_number"
    test_index_exists "service" "idx_service_fp"
    test_index_exists "vuln" "idx_vuln_cve"
    test_index_exists "country" "idx_country_cc"
    test_index_exists "asn" "idx_asn_number"
    echo ""

    # ========================================
    # Test Specific Seed Data Records
    # ========================================
    echo "=== Testing Specific Seed Data Records ==="
    test_record_exists "common_port" "label" "ssh"
    test_record_exists "common_port" "label" "http"
    test_record_exists "common_port" "label" "https"
    test_record_exists "common_port" "label" "mysql"
    test_record_exists "common_port" "label" "redis"
    test_record_exists "common_port" "label" "postgres"
    test_record_exists "common_port" "label" "mongodb"

    test_record_exists "country" "cc" "US"
    test_record_exists "country" "cc" "GB"
    test_record_exists "country" "cc" "FR"
    test_record_exists "country" "cc" "DE"
    test_record_exists "country" "cc" "JP"

    test_record_exists "cloud_region" "provider" "aws"
    test_record_exists "cloud_region" "provider" "gcp"
    test_record_exists "cloud_region" "provider" "azure"
    echo ""

    # ========================================
    # Summary
    # ========================================
    echo "========================================="
    echo "Verification Summary"
    echo "========================================="
    echo -e "${GREEN}Tests Passed: ${TESTS_PASSED}${NC}"
    echo -e "${RED}Tests Failed: ${TESTS_FAILED}${NC}"
    echo "========================================="

    if [ $TESTS_FAILED -eq 0 ]; then
        log_info "All tests passed! Schema and seed data are correctly configured."
        exit 0
    else
        log_fail "$TESTS_FAILED test(s) failed. Please review the schema and seed data."
        exit 1
    fi
}

# Run main function
main "$@"
