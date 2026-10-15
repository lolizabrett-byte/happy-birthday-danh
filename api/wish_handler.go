package main

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type createWishRequest struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Website string `json:"website"` // honeypot: real users leave it empty
}

func (h *handlers) listWishes(c echo.Context) error {
	wishes, err := h.store.ListWishes(c.Request().Context(), parseLimit(c, 50, 100))
	if err != nil {
		c.Logger().Errorf("list wishes: %v", err)
		return errInternal
	}
	return c.JSON(http.StatusOK, map[string][]Wish{"wishes": wishes})
}

func (h *handlers) createWish(c echo.Context) error {
	var req createWishRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	id, err := newID()
	if err != nil {
		return errInternal
	}
	w := Wish{ID: id, CreatedAt: nowUTC()}

	if w.Name, err = validateName(req.Name, maxWishName); err != nil {
		return badRequest(err)
	}
	if w.Message, err = validateMessage(req.Message); err != nil {
		return badRequest(err)
	}
	if strings.TrimSpace(req.Website) != "" { // bot: pretend success, store nothing
		return c.JSON(http.StatusCreated, w)
	}
	if err := h.store.AddWish(c.Request().Context(), w); err != nil {
		c.Logger().Errorf("add wish: %v", err)
		return errInternal
	}
	return c.JSON(http.StatusCreated, w)
}
