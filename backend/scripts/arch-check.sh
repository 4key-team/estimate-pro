#!/usr/bin/env bash
# arch-check.sh — архитектурные гейты для EstimatePro.
# Проверяет правила TDD+DDD+Clean Architecture, закреплённые в CLAUDE.md.
# Запускается локально (`just arch-check`) и в CI.
set -euo pipefail

cd "$(dirname "$0")/.."

violations=0

report_fail() {
    echo "  FAIL: $1"
    violations=$((violations + 1))
}

# --- Rule 1: domain — чистый, без инфры ---
echo "→ 1/4  domain/ не импортирует pgx/redis/chi/minio"
infra_in_domain=$(grep -rn --include='*.go' \
    -E 'github\.com/(jackc/pgx|redis/go-redis|go-chi|minio)' \
    internal/modules/*/domain/ 2>/dev/null || true)
if [ -n "$infra_in_domain" ]; then
    report_fail "инфра-либы в domain (должен быть pure):"
    echo "$infra_in_domain" | sed 's/^/    /'
fi

# --- Rule 2: нет cross-module импортов ---
echo "→ 2/4  modules/X не импортирует modules/Y"
for module_path in internal/modules/*/; do
    module=$(basename "$module_path")
    cross=$(grep -rn --include='*.go' \
        'github.com/VDV001/estimate-pro/backend/internal/modules/' \
        "$module_path" 2>/dev/null \
        | grep -v "internal/modules/$module/" || true)
    if [ -n "$cross" ]; then
        report_fail "cross-module импорт из $module:"
        echo "$cross" | sed 's/^/    /'
    fi
done

# --- Rule 3: handler не генерирует uuid/time (бизнес-логика) ---
echo "→ 3/4  handler не вызывает uuid.New()/time.Now() напрямую"
uuid_in_handler=$(grep -rn --include='*.go' --exclude='*_test.go' \
    -E 'uuid\.New\(\)|time\.Now\(\)' \
    internal/modules/*/handler/ 2>/dev/null || true)
if [ -n "$uuid_in_handler" ]; then
    filtered=$(echo "$uuid_in_handler" | grep -v '// arch:allow' || true)
    if [ -n "$filtered" ]; then
        report_fail "uuid.New()/time.Now() в handler (перенести в usecase, иначе пометить '// arch:allow TODO #X'):"
        echo "$filtered" | sed 's/^/    /'
    fi
fi

# --- Rule 4: handler не импортирует инфру напрямую ---
echo "→ 4/4  handler не импортирует pgx/redis/minio"
infra_in_handler=$(grep -rn --include='*.go' --exclude='*_test.go' \
    -E 'github\.com/(jackc/pgx|redis/go-redis|minio)' \
    internal/modules/*/handler/ 2>/dev/null || true)
if [ -n "$infra_in_handler" ]; then
    report_fail "инфра-импорт в handler (должен жить в repository):"
    echo "$infra_in_handler" | sed 's/^/    /'
fi

echo
if [ "$violations" -gt 0 ]; then
    echo "arch-check: FAIL ($violations violations)"
    exit 1
fi
echo "arch-check: PASS"
