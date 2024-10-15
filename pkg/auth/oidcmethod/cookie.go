/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package oidcmethod

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	theliverr "github.com/fidelity/theliv/pkg/err"
	log "github.com/fidelity/theliv/pkg/log"
	"github.com/go-chi/render"
)

const (
	maxValueLength = 3900
)

func splitCookie(key string, value string) []*http.Cookie {
	cookies := []*http.Cookie{}
	valueLength := len(value)
	numberOfChunks := int(math.Ceil(float64(valueLength) / float64(maxValueLength)))
	var end int
	for i, j := 0, 0; i < valueLength; i, j = i+maxValueLength, j+1 {
		end = i + maxValueLength
		if end > valueLength {
			end = valueLength
		}
		cookie := &http.Cookie{
			Path:     "/",
			Expires:  time.Now().Add(time.Hour),
			Secure:   true,
			HttpOnly: true,
		}
		if j == 0 && numberOfChunks == 1 {
			cookie.Name = key
			cookie.Value = value[i:end]
		} else if j == 0 {
			cookie.Name = key
			cookie.Value = fmt.Sprintf("%d:%s", numberOfChunks, value[i:end])
		} else {
			cookie.Name = fmt.Sprintf("%s-%d", key, j)
			cookie.Value = string(value[i:end])
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}

func joinCookies(ctx context.Context, key string, cookieList []*http.Cookie) (string, error) {
	cookies := make(map[string]string)
	for _, cookie := range cookieList {
		if !strings.HasPrefix(cookie.Name, key) {
			continue
		}
		cookies[cookie.Name] = cookie.Value
	}

	var sb strings.Builder
	var numOfChunks int
	var err error
	var token string
	var ok bool
	if token, ok = cookies[key]; !ok {
		msg := "failed to retrieve id_token from cookies"
		log.SWithContext(ctx).Warnf(msg)
		return "", theliverr.NewCommonError(ctx, 1, msg)
	}
	parts := strings.Split(token, ":")

	if len(parts) == 2 {
		if numOfChunks, err = strconv.Atoi(parts[0]); err != nil {
			return "", err
		}
		sb.WriteString(parts[1])
	} else if len(parts) == 1 {
		numOfChunks = 1
		sb.WriteString(parts[0])
	} else {
		log.SWithContext(ctx).Warn("invalid cookie %s")
		return "", fmt.Errorf("invalid cookie for key %s", key)
	}

	for i := 1; i < numOfChunks; i++ {
		sb.WriteString(cookies[fmt.Sprintf("%s-%d", key, i)])
	}
	return sb.String(), nil
}

func processError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(theliverr.GetStatusCode(err))
	render.JSON(w, r, err)
}
