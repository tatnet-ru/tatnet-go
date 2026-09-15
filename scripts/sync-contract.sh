#!/usr/bin/env bash
#
# Синхронизировать openapi/v1.json с публичным контрактом /v1.
#
#   scripts/sync-contract.sh           скачать и записать
#   scripts/sync-contract.sh --check   выйти с ошибкой, если разошлось
#
# Источник правды — ЖИВОЙ прод: документ там публикуется (V1_DOCS_ENABLED) и
# гейтится в самом api (tests/test_v1_openapi_snapshot.py), поэтому токен и
# доступ к приватному репозиторию api здесь не нужны.
set -euo pipefail
cd "$(dirname "$0")/.."

URL="${TATNET_OPENAPI_URL:-https://api.tatnet.ru/v1/openapi.json}"
LOCAL="openapi/v1.json"

mode="write"
[ "${1:-}" = "--check" ] && mode="check"

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

curl -fsS --max-time 30 "$URL" \
  | python3 -c 'import json,sys; json.dump(json.load(sys.stdin), sys.stdout, indent=2, sort_keys=True, ensure_ascii=False); print()' \
  > "$tmp"

if [ "$mode" = "check" ]; then
  if ! diff -q "$LOCAL" "$tmp" >/dev/null 2>&1; then
    echo "ERROR: $LOCAL разошёлся с $URL — публичный контракт изменился." >&2
    echo "Обновить и перегенерировать клиент:" >&2
    echo "  scripts/sync-contract.sh && scripts/generate.sh" >&2
    exit 1
  fi
  echo "OK: $LOCAL совпадает с $URL."
else
  cp "$tmp" "$LOCAL"
  echo "Обновлён $LOCAL из $URL."
fi
