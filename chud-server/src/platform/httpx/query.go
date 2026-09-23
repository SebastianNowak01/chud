package httpx

import (
	"fmt"
	"net/http"

	"github.com/sebnow/chud/platform/clock"
)

func DateRangeQuery(r *http.Request) (clock.Date, clock.Date, error) {
	from, err := dateQuery(r, "from")
	if err != nil {
		return clock.Date{}, clock.Date{}, err
	}
	to, err := dateQuery(r, "to")
	if err != nil {
		return clock.Date{}, clock.Date{}, err
	}
	return from, to, nil
}

func dateQuery(r *http.Request, key string) (clock.Date, error) {
	date, err := clock.ParseDate(r.URL.Query().Get(key))
	if err != nil {
		return clock.Date{}, fmt.Errorf("%s musi być datą RRRR-MM-DD", key)
	}
	return date, nil
}
