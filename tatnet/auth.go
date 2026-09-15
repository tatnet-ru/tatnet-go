// Package tatnet — сгенерированный клиент публичного API TatNet (/v1).
//
// Весь остальной код пакета порождается из openapi/v1.json и правке руками не
// подлежит; здесь — небольшая ручная обвязка, которой генератор не даёт:
// авторизация и базовый адрес.
//
//	c, err := tatnet.New(os.Getenv("TATNET_API_KEY"))
//	if err != nil { ... }
//	resp, err := c.GetAccountAccountGetWithResponse(ctx)
package tatnet

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// DefaultBaseURL — адрес публичного API. Совпадает с блоком servers в
// контракте; менять только вместе с ним.
const DefaultBaseURL = "https://api.tatnet.ru/v1"

// ErrEmptyAPIKey — пустой ключ отвергается сразу, а не превращается в
// заголовок «Bearer », на который сервер ответит 401 без объяснения.
var ErrEmptyAPIKey = errors.New("tatnet: пустой API-ключ")

// WithAPIKey добавляет к каждому запросу ключ: Authorization: Bearer tn_live_…
//
// Ключ принадлежит ОДНОМУ аккаунту и несёт политику, поэтому endpoint'ы не
// принимают account_id — аккаунт подразумевается ключом.
func WithAPIKey(key string) ClientOption {
	return WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		if strings.TrimSpace(key) == "" {
			return ErrEmptyAPIKey
		}
		req.Header.Set("Authorization", "Bearer "+key)
		return nil
	})
}

// New — клиент с типизированными ответами на боевом адресе.
//
// Отдельный конструктор, а не голый NewClientWithResponses, потому что иначе
// каждый потребитель повторял бы одно и то же: адрес, схему авторизации и
// проверку пустого ключа. Ровно это и разошлось в рукописных клиентах
// cloud-controller-manager и csi-driver.
func New(apiKey string, opts ...ClientOption) (*ClientWithResponses, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrEmptyAPIKey
	}
	all := append([]ClientOption{WithAPIKey(apiKey)}, opts...)
	return NewClientWithResponses(DefaultBaseURL, all...)
}
