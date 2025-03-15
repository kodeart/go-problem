package problem

import (
    "fmt"
    "net/http"
    "reflect"
    "strconv"
)

var cacheableStatuses = []int{
    0,
    http.StatusOK,
    http.StatusNoContent,
    http.StatusPartialContent,
    http.StatusMultipleChoices,
    http.StatusMovedPermanently,
    http.StatusNotFound,
    http.StatusMethodNotAllowed,
    http.StatusGone,
    http.StatusRequestURITooLong,
    http.StatusNotImplemented,
}

// statusCode converts various numerical types into int.
//
// [IMPORTANT]: built-in json.Unmarshaler converts numeric values
// into float64, so we need to convert status code back to int.
// If the status code value is a numeric string, it will be converted to int.
func (p Problem) statusCode(value any) (int, error) {
    switch v := reflect.ValueOf(value); v.Kind() {
    case reflect.String:
        i, err := strconv.Atoi(v.String())
        if err != nil {
            return 0, fmt.Errorf("invalid status type: %v", v.Kind())
        }
        return i, nil
    case reflect.Float32, reflect.Float64:
        return int(v.Float()), nil
    default:
        return 0, fmt.Errorf("invalid status type: %T", value)
    }
}

// setCacheControl sets the `Cache-Control` header
// if the status code is not in the cacheableStatuses list.
func setCacheControl(w http.ResponseWriter, status int) http.ResponseWriter {
    for _, s := range cacheableStatuses {
        if status == s {
            return w
        }
    }
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    return w
}
