# tatnet-go

Каталог тарифов ВМ: `VmsListPlansWithResponse` (`GET /v1/vms/plans`).
`VmsListPlansParams` принимает UUID региона `ClusterId` и назначение `Role`
(`vm`, `k8s`, `db`, `lb`). `JSON200.Plans` содержит UUID для `vm_plan_id`,
CPU, память в МиБ, ограничения диска и цены в рублях; `nil` у цены означает
отсутствие ставки. Требуется API-ключ и `vm:read`.

[tatnet.ru](https://tatnet.ru) · [панель](https://min.tatnet.ru) ·
[документация](https://docs.tatnet.ru) ·
[справочник API](https://api.tatnet.ru/v1/docs)

Go-клиент публичного API облака [TatNet](https://tatnet.ru) (`/v1`).
**Генерируется** из контракта — править `tatnet/client.gen.go` руками
бессмысленно, следующая генерация затрёт.

Ключ доступа создаётся в панели:
[min.tatnet.ru/api-keys](https://min.tatnet.ru/api-keys).

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

## Подписки ВМ

Для календарного месяца при создании используйте `PeriodMonths: &months`
с `months := 1`. `PeriodDays` поддерживает только `1`: тридцать дней не
являются месячной подпиской. Без обоих полей API создаёт почасовую подписку.

`VmsGetVmSubscriptionWithResponse` читает фактический период, цену и
автопродление. `VmsPatchVmSubscriptionWithResponse` меняет `PeriodType`
(`1h`, `1d`, `1m`, `3m`, `6m`, `12m`) и/или `AutoRenew` без пересоздания и
перезапуска ВМ. Повтор одинаковых настроек не списывает деньги. Смена периода
возвращает неиспользованную предоплату бонусом и оплачивает новый период;
просроченную подписку сначала нужно продлить. Автопродление само по себе
не оплачивает и не запускает остановленную ВМ.

## Требование к версии Go — минимальное

`go 1.24.0` — это пол зависимости `oapi-codegen/runtime`, а не удобство
автора. SDK не имеет права поднимать тулчейн у потребителей: директива
`go` в библиотеке становится нижней границей для всех, кто её подключает,
и потребитель с закреплённым тулчейном перестанет собираться.

## Версия генератора пиньена

`oapi-codegen v2.8.0`. Понижать нельзя: FastAPI отдаёт OpenAPI **3.1**, где
`exclusiveMinimum` — число, а не булево как в 3.0. На этом падают и v2.5.0, и
`ogen v1.24`; такое поле в контракте ровно одно, и его хватает, чтобы
отвергнуть весь документ.

## Ссылки

| | |
|---|---|
| Облако TatNet | [tatnet.ru](https://tatnet.ru) |
| Панель управления | [min.tatnet.ru](https://min.tatnet.ru) |
| Ключи доступа | [min.tatnet.ru/api-keys](https://min.tatnet.ru/api-keys) |
| Документация платформы | [docs.tatnet.ru](https://docs.tatnet.ru) |
| Справочник `/v1` | [api.tatnet.ru/v1/docs](https://api.tatnet.ru/v1/docs) |
| Контракт OpenAPI | [api.tatnet.ru/v1/openapi.json](https://api.tatnet.ru/v1/openapi.json) |
| Консольный клиент | [tatnet-ru/tatnet-cli](https://github.com/tatnet-ru/tatnet-cli) — `npx tatnet` |

## Лицензия

[Apache-2.0](LICENSE).


## Образы, доступные аккаунту

`VmsListImagesWithResponse(ctx, projectID, &tatnet.VmsListImagesParams{ClusterId: &regionID})`
возвращает каталог с учётом разрешений аккаунта API-ключа. Параметр account_id
не передаётся. Для выбора образа в регионе используйте `versions[].by_cluster[regionID]`:
`image_id` может указывать на более новую сборку, ещё не разложенную в регионе.
Отсутствие RED OS в каталоге означает, что нет доступной сборки; SDK не выдаёт разрешения.
Проверяйте HTTP status и JSON200: отказ доступа к проекту не является пустым каталогом.

## Реплики PostgreSQL

`PostgresChangeReadReplicasWithResponse` записывает число выделенных
асинхронных реплик (`ChangePgReadReplicasRequest.ReadReplicaCount`, 0–5).
`CreatePgClusterRequest.ReadReplicaCount` позволяет создать их вместе с кластером.
Проверяйте `ReadReplicasPending`, узлы и лаг в `PostgresGetClusterWithResponse`.
Reader endpoint возвращается как `ReadHost` / `ReadPort` (TCP/5433), независимо
от HA; primary не используется как fallback. Каждая реплика оплачивается по
тарифу узла. `VersionCapabilities` описывает возможности выбранного мажора.

## Зарезервированный IP балансировщика

`V1LBCreate.FloatingIpId` передаёт ID существующей доступной резервации
в том же аккаунте и регионе. `LoadBalancerResponse.FloatingIpId` и
`LoadBalancerDetailResponse.FloatingIpId` подтверждают привязку; `nil` означает
автоматически выделенный VIP. У Floating IP поля `LoadBalancerId` и
`LoadBalancerProjectId` описывают его использование балансировщиком.

Удаление LB сохраняет резервацию. Для повторного использования нужно дождаться
завершения регионального удаления, когда Floating IP вернётся в `available`.
Сервер должен поддерживать этот контракт (миграция 00298 и соответствующий API);
наличие поля в SDK само по себе не обновляет сервер.
