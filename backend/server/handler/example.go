package handler

import (
	"github.com/DaniloReisRocha/grid-tactics/backend/internal/utils/endpoint"
	"github.com/DaniloReisRocha/grid-tactics/backend/models"
	"github.com/DaniloReisRocha/grid-tactics/backend/server/middleware"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

type ExampleHandler struct {
	create models.ExampleCreate
	update models.ExampleUpdate
	list   models.ExampleList
	get    models.ExampleGet
	delete models.ExampleDelete
}

func (handler *ExampleHandler) Create() endpoint.APIEndpoint {
	return endpoint.NewPOSTRoute("/api/example", true, func(
		c *gin.Context,
		body models.ExampleDatabaseForm,
		_ endpoint.NoParams,
		_ endpoint.NoQuery,
	) (res endpoint.SingleItemData[models.ExampleDatabaseView], err error) {
		if _, ok := middleware.ValidateAccoID(c); !ok {
			return res, err
		}
		item, err := handler.create(body)
		if err != nil {
			return res, errors.Wrap(err, "Fail to create new example")
		}

		return endpoint.NewSingleItemData(*item, "create_example"), nil
	})
}

func (handler *ExampleHandler) Update() endpoint.APIEndpoint {
	return endpoint.NewPUTRoute("/api/example/:id", true, func(
		c *gin.Context,
		body models.ExampleDatabaseForm,
		params struct {
			ID uuid.UUID `uri:"id,string" binding:"required"`
		},
		_ endpoint.NoQuery,
	) (res endpoint.SingleItemData[models.ExampleDatabaseView], err error) {
		if _, ok := middleware.ValidateAccoID(c); !ok {
			return res, err
		}

		item, err := handler.update(params.ID, body)
		if err != nil {
			return res, errors.Wrap(err, "Fail to update example")
		}

		return endpoint.NewSingleItemData(*item, "update_example"), nil
	})
}

func (handler *ExampleHandler) Get() endpoint.APIEndpoint {
	return endpoint.NewGETRoute("/api/example/:id", true, func(
		c *gin.Context,
		_ endpoint.NoBody,
		params struct {
			ID uuid.UUID `uri:"id" binding:"required"`
		},
		_ endpoint.NoQuery,
	) (res endpoint.SingleItemData[models.ExampleDatabaseView], err error) {
		if _, ok := middleware.ValidateAccoID(c); !ok {
			return res, err
		}
		item, err := handler.get(params.ID)
		if err != nil {
			return res, errors.Wrap(err, "Fail to get example")
		}

		return endpoint.NewSingleItemData(*item, "get_example"), nil
	})
}

func (handler *ExampleHandler) List() endpoint.APIEndpoint {
	return endpoint.NewGETRoute("/api/example", true, func(
		c *gin.Context,
		_ endpoint.NoBody,
		_ endpoint.NoParams,
		query models.ExampleDatabaseFilter,
	) (res endpoint.CollectionItemData[models.ExampleDatabaseView], err error) {
		if _, ok := middleware.ValidateAccoID(c); !ok {
			return res, err
		}

		items, err := handler.list(query)
		if err != nil {
			return res, errors.Wrap(err, "Fail to list example")
		}

		return endpoint.NewCollectionItemData(items, "list_example"), nil
	})
}

func (handler *ExampleHandler) Delete() endpoint.APIEndpoint {
	return endpoint.NewDELETERoute("/api/example/:id", true, func(
		c *gin.Context,
		_ endpoint.NoBody,
		params struct {
			ID uuid.UUID `uri:"id" binding:"required"`
		},
		_ endpoint.NoQuery,
	) (res endpoint.SingleItemData[models.ExampleDatabaseView], err error) {
		if _, ok := middleware.ValidateAccoID(c); !ok {
			return res, err
		}

		item, err := handler.delete(params.ID)
		if err != nil {
			return res, errors.Wrap(err, "Fail to delete example")
		}

		return endpoint.NewSingleItemData(*item, "update_example"), nil
	})
}
