#!/usr/bin/env bash
#
# Перегенерировать клиент из КОММИТНУТОГО контракта openapi/v1.json.
#
#   scripts/generate.sh           перезаписать tatnet/client.gen.go
#   scripts/generate.sh --check   выйти с ошибкой, если файл разошёлся
#
# Контракт сюда копируется отдельно (scripts/sync-contract.sh) — так генерация
# не ходит в сеть и воспроизводима на любой машине и в CI.
set -euo pipefail
cd "$(dirname "$0")/.."

# ⚠ Версия пинится и понижать её нельзя: FastAPI отдаёт OpenAPI 3.1, где
# exclusiveMinimum — ЧИСЛО, а не булево как в 3.0. v2.5.0 на этом падает
# («cannot unmarshal number into field Schema.exclusiveMinimum of type bool»),
# как и ogen v1.24. В документе такое поле ровно одно — и его хватает, чтобы
# отвергнуть весь контракт.
GEN="github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.0"
OUT="tatnet/client.gen.go"

mode="write"
[ "${1:-}" = "--check" ] && mode="check"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

go run "$GEN" -config oapi-codegen.yaml -o "$tmp/client.gen.go" openapi/v1.json

if [ "$mode" = "check" ]; then
  if ! diff -q "$OUT" "$tmp/client.gen.go" >/dev/null 2>&1; then
    echo "ERROR: $OUT разошёлся с контрактом openapi/v1.json." >&2
    echo "Файл ГЕНЕРИРУЕТСЯ — править руками нельзя. Перегенерировать:" >&2
    echo "  scripts/generate.sh" >&2
    diff -u "$OUT" "$tmp/client.gen.go" | head -40 >&2
    exit 1
  fi
  echo "OK: $OUT соответствует openapi/v1.json."
else
  mkdir -p "$(dirname "$OUT")"
  cp "$tmp/client.gen.go" "$OUT"
  echo "Записан $OUT ($(wc -l < "$OUT" | tr -d ' ') строк)."
fi
