package entries

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/clock"
	"github.com/sebnow/chud/platform/constants"
	"github.com/sebnow/chud/platform/httpx"
)

const (
	maxRequestSize = MaxFileCount*MaxFileSize + 1<<20
	maxFormMemory  = 8 << 20
	maxUploads     = 3
)

type EntryAPIController struct {
	service IEntryService
	uploads chan struct{}
}

func NewEntryAPIController(service IEntryService) EntryAPIController {
	return EntryAPIController{
		service: service,
		uploads: make(chan struct{}, maxUploads),
	}
}

func (c *EntryAPIController) GetEntriesByActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	entries, err := c.service.GetEntriesByActivity(ctx, id)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, entries)
}

func (c *EntryAPIController) CreateEntryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	select {
	case c.uploads <- struct{}{}:
		defer func() { <-c.uploads }()
	case <-ctx.Done():
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	if err := r.ParseMultipartForm(maxFormMemory); err != nil {
		if _, tooLarge := errors.AsType[*http.MaxBytesError](err); tooLarge {
			httpx.RespondError(ctx, w, http.StatusRequestEntityTooLarge, errors.New("pliki są za duże"))
			return
		}
		httpx.RespondError(ctx, w, http.StatusBadRequest, errors.New("nieprawidłowy formularz"))
		return
	}
	defer r.MultipartForm.RemoveAll()

	payload, err := parseCreateEntryForm(r.MultipartForm)
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	entry, svcErr := c.service.CreateEntry(ctx, id, claims.UserID, payload)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusCreated, entry)
}

func (c *EntryAPIController) GetMediaByEntryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	media, err := c.service.GetMediaByEntry(ctx, id)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, media)
}

func (c *EntryAPIController) GetMediaHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	media, err := c.service.GetMedia(ctx, id)
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	h := w.Header()
	if IsAllowedMediaType(media.ContentType) {
		h.Set(constants.HTTPHeaderContentType, media.ContentType)
		h.Set(constants.HTTPHeaderContentDisposition, "inline")
	} else {
		h.Set(constants.HTTPHeaderContentType, "application/octet-stream")
		h.Set(constants.HTTPHeaderContentDisposition, "attachment")
	}
	h.Set(constants.HTTPHeaderContentSecurityPolicy, httpx.MediaContentSecurityPolicy)
	h.Set(constants.HTTPHeaderCacheControl, "private, max-age=31536000, immutable")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(media.Data))
}

func (c *EntryAPIController) GetEntriesInRangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	from, to, err := httpx.DateRangeQuery(r)
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	entries, svcErr := c.service.GetEntriesInRange(ctx, from, to)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, entries)
}

func (c *EntryAPIController) GetMyEntriesInRangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	from, to, err := httpx.DateRangeQuery(r)
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	entries, svcErr := c.service.GetUserEntriesInRange(ctx, claims.UserID, from, to)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, entries)
}

func (c *EntryAPIController) GetUserEntriesInRangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, ok := httpx.PathUUIDOrRespond(w, r, "id")
	if !ok {
		return
	}
	from, to, err := httpx.DateRangeQuery(r)
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	entries, svcErr := c.service.GetUserEntriesInRange(ctx, id, from, to)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, entries)
}

func parseCreateEntryForm(form *multipart.Form) (CreateEntryPayload, error) {
	value := func(key string) string {
		if values := form.Value[key]; len(values) > 0 {
			return strings.TrimSpace(values[0])
		}
		return ""
	}

	payload := CreateEntryPayload{
		Description: value("description"),
		Excused:     value("excused") == "true",
	}

	if v := value("occurredAt"); v != "" {
		occurredAt, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return payload, fmt.Errorf("occurredAt musi być znacznikiem czasu RFC 3339")
		}
		payload.OccurredAt = occurredAt
	}
	if v := value("planId"); v != "" {
		if uuid.Validate(v) != nil {
			return payload, errors.New("nie znaleziono: plan")
		}
		payload.PlanID = &v
	}
	if v := value("scheduledFor"); v != "" {
		scheduledFor, err := clock.ParseDate(v)
		if err != nil {
			return payload, fmt.Errorf("scheduledFor musi być datą RRRR-MM-DD")
		}
		payload.ScheduledFor = &scheduledFor
	}

	headers := form.File["files"]
	if len(headers) > MaxFileCount {
		return payload, fmt.Errorf("maksymalnie %d plików na wpis", MaxFileCount)
	}
	for _, header := range headers {
		file, err := readFile(header)
		if err != nil {
			return payload, err
		}
		payload.Files = append(payload.Files, file)
	}

	return payload, nil
}

func readFile(header *multipart.FileHeader) (Media, error) {
	if header.Size > MaxFileSize {
		return Media{}, fmt.Errorf("plik %s ma więcej niż 10 MB", header.Filename)
	}

	f, err := header.Open()
	if err != nil {
		return Media{}, fmt.Errorf("nie udało się odczytać pliku %s", header.Filename)
	}
	defer f.Close()

	head := make([]byte, sniffLength)
	n, err := io.ReadFull(f, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return Media{}, fmt.Errorf("nie udało się odczytać pliku %s", header.Filename)
	}
	contentType := DetectMediaType(head[:n])
	if !IsAllowedMediaType(contentType) {
		return Media{}, fmt.Errorf("plik %s: %s", header.Filename, unsupportedMediaMessage)
	}

	rest, err := io.ReadAll(io.LimitReader(f, MaxFileSize))
	if err != nil {
		return Media{}, fmt.Errorf("nie udało się odczytać pliku %s", header.Filename)
	}
	return Media{ContentType: contentType, Data: append(head[:n], rest...)}, nil
}
