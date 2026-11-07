package endpoint

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type dataDetail struct {
	// The kind property serves as a guide to what type of information this particular object stores
	Kind string `json:"kind" example:"resource"`
}

type SingleItemData[T any] struct {
	dataDetail
	Item T `json:"item"`
}

type CollectionItemData[T any] struct {
	dataDetail
	Items []T `json:"items"`
	collectionDetail
}

type collectionDetail struct {
	// The number of items in this result set
	CurrentItemCount int64 `json:"currentItemCount" example:"1"`
}

func NewSingleItemData[T any](item T, kind string) SingleItemData[T] {
	return SingleItemData[T]{
		dataDetail: dataDetail{
			Kind: kind,
		},
		Item: item,
	}
}

func NewCollectionItemData[T any](items []T, kind string) CollectionItemData[T] {
	return CollectionItemData[T]{
		dataDetail: dataDetail{
			Kind: kind,
		},
		Items: items,
		collectionDetail: collectionDetail{
			CurrentItemCount: int64(len(items)),
		},
	}
}

type NoBody = struct{}

type NoParams = struct{}

type NoQuery = struct{}

type HandlerWithTypes[B any, P any, Q any, R any] func(c *gin.Context, body B, params P, query Q) (R, error)
type WebSocketHandler func(w http.ResponseWriter, r *http.Request)

type APIEndpoint struct {
	Protected   bool
	NewEndpoint func() (string, string, gin.HandlerFunc)
	MakeZod     func() (*Zod, error)
}

func NewPOSTRoute[Body any, Params any, Query any, Response any](path string, protected bool, handler HandlerWithTypes[Body, Params, Query, Response]) APIEndpoint {
	return NewRoute(http.MethodPost, path, protected, handler)
}

func NewGETRoute[Params any, Query any, Response any](path string, protected bool, handler HandlerWithTypes[struct{}, Params, Query, Response]) APIEndpoint {
	return NewRoute(http.MethodGet, path, protected, handler)
}

func NewPUTRoute[Body any, Params any, Query any, Response any](path string, protected bool, handler HandlerWithTypes[Body, Params, Query, Response]) APIEndpoint {
	return NewRoute(http.MethodPut, path, protected, handler)
}

func NewPATCHRoute[Body any, Params any, Query any, Response any](path string, protected bool, handler HandlerWithTypes[Body, Params, Query, Response]) APIEndpoint {
	return NewRoute(http.MethodPatch, path, protected, handler)
}

func NewDELETERoute[Params any, Query any, Response any](path string, protected bool, handler HandlerWithTypes[struct{}, Params, Query, Response]) APIEndpoint {
	return NewRoute(http.MethodDelete, path, protected, handler)
}

func NewRoute[B any, P any, Q any, R any](method, path string, protected bool, handler HandlerWithTypes[B, P, Q, R]) APIEndpoint {
	return APIEndpoint{
		Protected:   protected,
		NewEndpoint: newEndpoint(method, path, handler),
		MakeZod: func() (*Zod, error) {
			return MakeZod[B, P, Q, R](method, path)
		},
	}
}

func newEndpoint[B any, P any, Q any, R any](method, path string, handler HandlerWithTypes[B, P, Q, R]) func() (string, string, gin.HandlerFunc) {
	return func() (string, string, gin.HandlerFunc) {
		return method, path, func(c *gin.Context) {
			var body B
			var query Q

			if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
				if err := c.ShouldBindJSON(&body); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
					return
				}
			}

			params, err := getParams[P](c)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "URI parameters", "details": err.Error()})
				return
			}

			if err := c.ShouldBindQuery(&query); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters", "details": err.Error()})
				return
			}

			if r, err := handler(c, body, params, query); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			} else {
				c.JSON(http.StatusOK, r)
			}
		}
	}
}

func NewWebSocketRoute(path string, handler WebSocketHandler) APIEndpoint {
	return APIEndpoint{
		Protected:   false,
		NewEndpoint: newWebSocketEndpoint(path, handler),
	}
}

func newWebSocketEndpoint(path string, handler WebSocketHandler) func() (string, string, gin.HandlerFunc) {
	return func() (string, string, gin.HandlerFunc) {
		return http.MethodGet, path, func(c *gin.Context) {
			handler(c.Writer, c.Request)
		}
	}
}

func getParams[P any](c *gin.Context) (params P, err error) {
	p := map[string]string{}
	if err = c.ShouldBindUri(&p); err != nil {
		return params, errors.Wrap(err, "bind uri")
	}
	params, err = mapToStruct(p, *new(P))
	if err != nil {
		return params, errors.Wrap(err, "map to struct")
	}
	return params, nil
}

func mapToStruct[K any, M ~map[string]K, S any](in M, out S) (S, error) {
	params, err := json.Marshal(in)
	if err != nil {
		return out, err
	}

	err = json.Unmarshal((params), &out)
	if err != nil {
		return out, err
	}
	return out, err
}
