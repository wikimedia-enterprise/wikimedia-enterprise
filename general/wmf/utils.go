package wmf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func getResponseError(rsp *Response) error {
	if rsp.Error != nil {
		dta, err := json.Marshal(rsp.Error)

		if err != nil {
			return err
		}

		return errors.New(string(dta))
	}

	return nil
}

func getRetryAfterValue(res *http.Response, def time.Duration) (time.Duration, error) {
	for _, val := range res.Header["Retry-After"] {
		if _, err := time.Parse(http.TimeFormat, val); err == nil {
			continue
		}

		rvl, err := strconv.Atoi(val)

		if err != nil {
			return def, err
		}

		return time.Second * time.Duration(rvl), nil
	}

	return def, nil
}

func getErrorString(res *http.Response) (string, error) {
	dta, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", res.Status, strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '\n' {
			return -1
		}

		return r
	}, string(dta))), nil
}
