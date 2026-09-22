package entries

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/sebnow/chud/platform/auth"
	"github.com/sebnow/chud/platform/constants"
	"github.com/sebnow/chud/platform/httpx"
)

const (
	maxRequestSize = MaxFileCount*MaxFileSize + 1<<20
	maxFormMemory  = 32 << 20
)

type EntryAPIController struct {
	service IEntryService
}

func NewEntryAPIController(service IEntryService) EntryAPIController {
	return EntryAPIController{
		service: service,
	}
}

func (c *EntryAPIController) GetEntriesByActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	entries, err := c.service.GetEntriesByActivity(ctx, r.PathValue("id"))
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, entries)
}

// CreateEntryHandler accepts multipart/form-data with the fields
// description, occurredAt (RFC 3339), planId, scheduledFor (YYYY-MM-DD), excused ("true") and files.
func (c *EntryAPIController) CreateEntryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)
	if err := r.ParseMultipartForm(maxFormMemory); err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, fmt.Errorf("invalid form: %w", err))
		return
	}
	defer r.MultipartForm.RemoveAll()

	payload, err := parseCreateEntryForm(r.MultipartForm)
	if err != nil {
		httpx.RespondError(ctx, w, http.StatusBadRequest, err)
		return
	}

	entry, svcErr := c.service.CreateEntry(ctx, r.PathValue("id"), claims.UserID, payload)
	if svcErr != nil {
		httpx.RespondError(ctx, w, svcErr.Code, svcErr.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusCreated, entry)
}

func (c *EntryAPIController) GetMediaByEntryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	media, err := c.service.GetMediaByEntry(ctx, r.PathValue("id"))
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	httpx.RespondJSON(ctx, w, http.StatusOK, media)
}

func (c *EntryAPIController) GetMediaHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	media, err := c.service.GetMedia(ctx, r.PathValue("id"))
	if err != nil {
		httpx.RespondError(ctx, w, err.Code, err.Err)
		return
	}

	// Media never changes once uploaded.
	w.Header().Set(constants.HTTPHeaderContentType, media.ContentType)
	w.Header().Set(constants.HTTPHeaderCacheControl, "private, max-age=31536000, immutable")
	_, _ = w.Write(media.Data)
}

// GetEntriesInRangeHandler returns everyone's entries in [from, to); both are RFC 3339 query params.
func (c *EntryAPIController) GetEntriesInRangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	from, to, err := parseRange(r)
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

// GetMyEntriesInRangeHandler returns the current user's entries in [from, to).
func (c *EntryAPIController) GetMyEntriesInRangeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ExtractUserOrRespond(ctx, w, r)
	if !ok {
		return
	}

	from, to, err := parseRange(r)
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

func parseRange(r *http.Request) (time.Time, time.Time, error) {
	from, err := time.Parse(time.RFC3339, r.URL.Query().Get("from"))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("from must be an RFC 3339 timestamp")
	}
	to, err := time.Parse(time.RFC3339, r.URL.Query().Get("to"))
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("to must be an RFC 3339 timestamp")
	}
	return from.UTC(), to.UTC(), nil
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
			return payload, fmt.Errorf("occurredAt must be an RFC 3339 timestamp")
		}
		payload.OccurredAt = occurredAt
	}
	if v := value("planId"); v != "" {
		payload.PlanID = &v
	}
	if v := value("scheduledFor"); v != "" {
		scheduledFor, err := time.Parse(time.DateOnly, v)
		if err != nil {
			return payload, fmt.Errorf("scheduledFor must be a YYYY-MM-DD date")
		}
		payload.ScheduledFor = &scheduledFor
	}

	for _, header := range form.File["files"] {
		file, err := readFile(header)
		if err != nil {
			return payload, err
		}
		payload.Files = append(payload.Files, file)
	}

	return payload, nil
}

// readFile reads an uploaded file, detecting its type from the content and falling back to the browser's.
func readFile(header *multipart.FileHeader) (Media, error) {
	if header.Size > MaxFileSize {
		return Media{}, fmt.Errorf("%s is larger than 10 MB", header.Filename)
	}

	f, err := header.Open()
	if err != nil {
		return Media{}, fmt.Errorf("failed to read %s: %w", header.Filename, err)
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return Media{}, fmt.Errorf("failed to read %s: %w", header.Filename, err)
	}

	contentType := http.DetectContentType(data)
	if contentType == "application/octet-stream" {
		contentType = header.Header.Get(constants.HTTPHeaderContentType)
	}
	return Media{ContentType: contentType, Data: data}, nil
}
