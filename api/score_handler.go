package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

type createScoreRequest struct {
	Name  string `json:"name"`
	Score *int   `json:"score"` // pointer: a missing field must be rejected, not stored as 0
}

func (h *handlers) listScores(c echo.Context) error {
	scores, err := h.store.TopScores(c.Request().Context(), parseLimit(c, 10, 50))
	if err != nil {
		c.Logger().Errorf("top scores: %v", err)
		return errInternal
	}
	return c.JSON(http.StatusOK, map[string][]Score{"scores": scores})
}

func (h *handlers) createScore(c echo.Context) error {
	var req createScoreRequest
	if err := bindJSON(c, &req); err != nil {
		return err
	}
	name, err := validateName(req.Name, maxScoreName)
	if err != nil {
		return badRequest(err)
	}
	if req.Score == nil {
		return badRequest(errors.New("score is required"))
	}
	if err := validateScore(*req.Score); err != nil {
		return badRequest(err)
	}
	id, err := newID()
	if err != nil {
		return errInternal
	}
	s := Score{ID: id, Name: name, Score: *req.Score, CreatedAt: nowUTC()}
	if err := h.store.AddScore(c.Request().Context(), s); err != nil {
		c.Logger().Errorf("add score: %v", err)
		return errInternal
	}
	return c.JSON(http.StatusCreated, s)
}
