package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/lib/logger/sl"
	"url-shortener/internal/service"

	"github.com/go-chi/chi/v5"
)

var (
	ErrInternal    = errors.New("internal server error")
	ErrURLNotFound = errors.New("URL not found")
	ErrInvalidURL  = errors.New("invalid URL format")
	ErrInvalidJSON = errors.New("invalid JSON")
)

type URLResponse struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	Code        string `json:"code"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type URLHandler struct {
	service service.Service
	logger  *slog.Logger
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type UpdateURLRequest struct {
	URL             *string `json:"url"`
	RefreshShortURL bool    `json:"refresh_short_url"`
}

func NewURLHandler(service service.Service, logger *slog.Logger) *URLHandler {
	return &URLHandler{
		service: service,
		logger:  logger,
	}
}

func (h *URLHandler) Save(w http.ResponseWriter, r *http.Request) {
	var req CreateURLRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidJSON.Error())
		return
	}

	res, err := h.service.SaveURL(r.Context(), req.URL)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyURL), errors.Is(err, service.ErrInvalidURL):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		default:
			h.logger.Error("request_failed_save", sl.Err(err))
			writeError(w, http.StatusInternalServerError, ErrInternal.Error())
			return
		}
	}

	resp := URLResponse{
		OriginalURL: res.OriginalURL,
		ShortURL:    res.ShortURL,
		Code:        res.Code,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *URLHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if code == "" {
		writeError(w, http.StatusBadRequest, ErrInvalidURL.Error())
		return
	}

	res, err := h.service.GetURLByCode(r.Context(), code)

	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			writeError(w, http.StatusNotFound, ErrURLNotFound.Error())
			return
		}

		h.logger.Error("request_failed_redirect", sl.Err(err))
		writeError(w, http.StatusInternalServerError, ErrInternal.Error())
		return
	}

	http.Redirect(w, r, res.OriginalURL, http.StatusFound)
}

func (h *URLHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	if code == "" {
		writeError(w, http.StatusBadRequest, ErrURLNotFound.Error())
		return
	}

	res, err := h.service.GetURLByCode(r.Context(), code)

	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			writeError(w, http.StatusNotFound, ErrURLNotFound.Error())
			return
		}

		h.logger.Error("request_failed_get", sl.Err(err))
		writeError(w, http.StatusInternalServerError, ErrInternal.Error())
		return
	}

	resp := URLResponse{
		OriginalURL: res.OriginalURL,
		ShortURL:    res.ShortURL,
		Code:        res.Code,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *URLHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateURLRequest

	code := chi.URLParam(r, "code")

	if code == "" {
		writeError(w, http.StatusBadRequest, ErrURLNotFound.Error())
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, ErrInvalidJSON.Error())
		return
	}

	url, err := h.service.UpdateURL(r.Context(), req.URL, req.RefreshShortURL, code)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidURL):
			writeError(w, http.StatusBadRequest, service.ErrInvalidURL.Error())
			return
		case errors.Is(err, service.ErrURLNotFound):
			writeError(w, http.StatusNotFound, ErrURLNotFound.Error())
			return
		case errors.Is(err, service.ErrNothingToUpdate):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, service.ErrCodeGenerationFailed):
			writeError(w, http.StatusInternalServerError, ErrInternal.Error())
			return
		case errors.Is(err, service.ErrURLAlreadyExists):
			writeError(w, http.StatusConflict, ErrInternal.Error())
			return
		default:
			h.logger.Error("request_failed_update", sl.Err(err))
			writeError(w, http.StatusInternalServerError, ErrInternal.Error())
			return
		}
	}

	resp := URLResponse{
		OriginalURL: url.OriginalURL,
		Code:        url.Code,
		ShortURL:    url.ShortURL,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *URLHandler) Delete(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	err := h.service.DeleteURL(r.Context(), code)

	if err != nil {
		if errors.Is(err, service.ErrURLNotFound) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.logger.Error("request_failed_delete", sl.Err(err))
		writeError(w, http.StatusInternalServerError, ErrInternal.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
