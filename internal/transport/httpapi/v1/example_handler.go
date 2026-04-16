package v1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"ai-go-chi-starter/internal/service/example"
	"ai-go-chi-starter/internal/service/shared"
	"ai-go-chi-starter/internal/transport/httpapi/httpx"
)

type ExampleService interface {
	Create(ctx context.Context, input example.CreateInput) (example.Example, error)
	Get(ctx context.Context, id string) (example.Example, error)
	List(ctx context.Context) ([]example.Example, error)
}

type ExampleHandler struct {
	service ExampleService
}

type createExampleRequest struct {
	Name string `json:"name"`
}

type exampleResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewExampleHandler(service ExampleService) *ExampleHandler {
	return &ExampleHandler{service: service}
}

func (h *ExampleHandler) Create(w http.ResponseWriter, req *http.Request) {
	if h.service == nil {
		httpx.WriteRequestError(
			w,
			req,
			http.StatusInternalServerError,
			shared.CodeInternal,
			"service is not configured",
			false,
		)
		return
	}
	var body createExampleRequest
	if err := httpx.DecodeRequestJSON(req, &body); err != nil {
		writeDecodeError(w, req, err)
		return
	}
	item, err := h.service.Create(req.Context(), example.CreateInput{Name: body.Name})
	if err != nil {
		httpx.WriteRequestDomainError(w, req, err)
		return
	}
	httpx.WriteEnvelope(w, http.StatusCreated, httpx.RequestID(req), toExampleResponse(item))
}

func writeDecodeError(w http.ResponseWriter, req *http.Request, err error) {
	var maxBytesErr *http.MaxBytesError
	switch {
	case errors.As(err, &maxBytesErr):
		httpx.WriteRequestError(
			w,
			req,
			http.StatusRequestEntityTooLarge,
			shared.CodeInvalidArgument,
			"request body is too large",
			false,
		)
	case errors.Is(err, httpx.ErrUnsupportedContentType):
		httpx.WriteRequestError(
			w,
			req,
			http.StatusUnsupportedMediaType,
			shared.CodeInvalidArgument,
			"Content-Type must be application/json",
			false,
		)
	case errors.Is(err, io.EOF):
		httpx.WriteRequestError(
			w,
			req,
			http.StatusBadRequest,
			shared.CodeInvalidArgument,
			"request body is empty",
			false,
		)
	case errors.Is(err, httpx.ErrMultipleJSONDocuments):
		httpx.WriteRequestError(
			w,
			req,
			http.StatusBadRequest,
			shared.CodeInvalidArgument,
			"request body must contain a single JSON document",
			false,
		)
	default:
		httpx.WriteRequestError(
			w,
			req,
			http.StatusBadRequest,
			shared.CodeInvalidArgument,
			"invalid JSON body",
			false,
		)
	}
}

func (h *ExampleHandler) Get(w http.ResponseWriter, req *http.Request) {
	if h.service == nil {
		httpx.WriteRequestError(
			w,
			req,
			http.StatusInternalServerError,
			shared.CodeInternal,
			"service is not configured",
			false,
		)
		return
	}
	item, err := h.service.Get(req.Context(), chi.URLParam(req, "id"))
	if err != nil {
		httpx.WriteRequestDomainError(w, req, err)
		return
	}
	httpx.WriteEnvelope(w, http.StatusOK, httpx.RequestID(req), toExampleResponse(item))
}

func (h *ExampleHandler) List(w http.ResponseWriter, req *http.Request) {
	if h.service == nil {
		httpx.WriteRequestError(
			w,
			req,
			http.StatusInternalServerError,
			shared.CodeInternal,
			"service is not configured",
			false,
		)
		return
	}
	items, err := h.service.List(req.Context())
	if err != nil {
		httpx.WriteRequestDomainError(w, req, err)
		return
	}
	response := make([]exampleResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toExampleResponse(item))
	}
	httpx.WriteEnvelope(w, http.StatusOK, httpx.RequestID(req), response)
}

func toExampleResponse(item example.Example) exampleResponse {
	return exampleResponse{
		ID:        item.ID,
		Name:      item.Name,
		CreatedAt: item.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
