package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/gommon/log"

	"github.com/knqyf263/boltwiz/modules/database/repository"

	"github.com/knqyf263/boltwiz/utils"

	"github.com/knqyf263/boltwiz/modules/database/model"

	"github.com/labstack/echo/v4"
)

type Handlers struct {
	repo *repository.Repository
}

func NewHandlers(repo *repository.Repository) *Handlers {
	return &Handlers{repo: repo}
}

func (h *Handlers) SayHello(c echo.Context) error {
	return c.String(200, "Hello from the other side")
}

func (h *Handlers) ListElement(c echo.Context) error {
	all, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var reqBody model.ListElemReqBody
	err = json.Unmarshal(all, &reqBody)
	if err != nil {
		return err
	}
	pageSize := utils.ParseInt(c.QueryParam("page_size"))
	pageNum := utils.ParseInt(c.QueryParam("page"))
	searchKey := c.QueryParam("key")
	reqBody.PageSize = pageSize
	reqBody.Page = pageNum
	reqBody.SearchKey = searchKey
	resp, err := h.repo.ListElement(reqBody)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed Listing element: %v", err))
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handlers) AddBucket(c echo.Context) error {
	all, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var reqBody model.BucketsToAdd
	err = json.Unmarshal(all, &reqBody)
	if err != nil {
		return err
	}
	err = h.repo.AddBuckets(reqBody)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed Adding bucket/s: %v", err))
	}
	return c.JSON(http.StatusOK, "Buckets added successfully")
}

func (h *Handlers) AddPairs(c echo.Context) error {
	all, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var reqBody model.PairsToAdd
	err = json.Unmarshal(all, &reqBody)
	if err != nil {
		return err
	}
	err = h.repo.AddPairs(reqBody)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed Adding pair/s: %v", err))
	}
	return c.JSON(http.StatusOK, "Pairs added successfully")
}

func (h *Handlers) DeleteElement(c echo.Context) error {
	all, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var reqBody model.ItemToDelete
	err = json.Unmarshal(all, &reqBody)
	if err != nil {
		return err
	}
	err = h.repo.DeleteElement(reqBody)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed Deleting element : %v", err))
	}
	return c.JSON(http.StatusOK, "Deleted successfully")
}
func (h *Handlers) RenameElement(c echo.Context) error {
	all, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var reqBody model.ItemToRename
	err = json.Unmarshal(all, &reqBody)
	if err != nil {
		return err
	}
	err = h.repo.RenameElement(reqBody)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed renaming element : %v", err))
	}
	return c.JSON(http.StatusOK, "Renamed successfully")
}
func (h *Handlers) UpdatePairValue(c echo.Context) error {
	all, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	var reqBody model.ItemToUpdate
	err = json.Unmarshal(all, &reqBody)
	if err != nil {
		return err
	}
	err = h.repo.UpdatePairValue(reqBody)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed updating pair value : %v", err))
	}
	return c.JSON(http.StatusOK, "Updated pair value successfully")
}
