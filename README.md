# tatnet-go

Go-клиент публичного API TatNet (`/v1`). **Генерируется** из контракта —
править `tatnet/client.gen.go` руками бессмысленно, следующая генерация затрёт.

```go
import "github.com/tatnet-ru/tatnet-go/tatnet"

c, err := tatnet.New(os.Getenv("TATNET_API_KEY"))
if err != nil { return err }

res, err := c.PostgresListClustersWithResponse(ctx, projectID, nil)
if err != nil { return err }
if res.JSON200 == nil {
    return fmt.Errorf("список кластеров: %s", res.Status())
}
for _, cl := range res.JSON200.Data {
    fmt.Println(cl.Name, cl.Status)
}
```

Ключ — `tn_live_…`, выпускается в панели (Настройки → API-ключи). Он
принадлежит одному аккаунту и несёт политику, поэтому ручки не принимают
`account_id`: аккаунт подразумевается ключом.

## Откуда берётся код

```
openapi/v1.json        контракт, копия опубликованного
  ↓ scripts/generate.sh
tatnet/client.gen.go   модели + клиент (oapi-codegen)
tatnet/auth.go         ручная обвязка: адрес и авторизация
```

Две команды, обе с режимом `--check` для CI:

```
scripts/sync-contract.sh     обновить openapi/v1.json с https://api.tatnet.ru/v1/openapi.json
scripts/generate.sh          перегенерировать клиент из openapi/v1.json
```

Разделение намеренное: генерация не ходит в сеть, поэтому воспроизводима на
любой машине и в CI. Сверка с живым контрактом — отдельный шаг, который делают
осознанно.

## Что держит CI

- `scripts/generate.sh --check` — закоммиченный клиент соответствует
  закоммиченному контракту. Правка руками или забытая перегенерация роняют
  сборку.
- `go build` / `go vet` / `go test`.

Контракт со своей стороны гейтится в api (`tests/test_v1_openapi_snapshot.py`):
он же проверяет то, на чём стоит кодогенерация — уникальность `operationId`,
наличие тела у каждого 2xx и отсутствие столкновений имён операций с именами
схем.

## Требование к версии Go — минимальное

`go 1.24.0` — это пол зависимости `oapi-codegen/runtime`, а не удобство
автора. SDK не имеет права поднимать тулчейн у потребителей: директива
`go` в библиотеке становится нижней границей для всех, кто её подключает.
Так и вышло — клиент с `go 1.26.2` поднял её у cloud-controller-manager,
чей Dockerfile собирает на 1.24 с `GOTOOLCHAIN=local`, и сборка образа
встала на `go mod download`.

## Версия генератора пиньена

`oapi-codegen v2.8.0`. Понижать нельзя: FastAPI отдаёт OpenAPI **3.1**, где
`exclusiveMinimum` — число, а не булево как в 3.0. На этом падают и v2.5.0, и
`ogen v1.24`; такое поле в контракте ровно одно, и его хватает, чтобы
отвергнуть весь документ.
