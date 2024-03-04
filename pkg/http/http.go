/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package http

import (
	"time"

	resty "github.com/go-resty/resty/v2"
)

func NewRetryClient(timeOut int, retryCount int, waitTime int, totalWaitTime int) *resty.Client {
	return newClient(timeOut).
		SetRetryCount(retryCount).
		SetRetryWaitTime(time.Duration(waitTime) * time.Second).
		SetRetryMaxWaitTime(time.Duration(totalWaitTime) * time.Second)
}

func NewRetryReq(timeOut int, retryCount int, waitTime int, totalWaitTime int) *resty.Request {
	return NewRetryClient(timeOut, retryCount, waitTime, totalWaitTime).
		R()
}

func newClient(timeout int) *resty.Client {
	return resty.New().SetTimeout(time.Second * time.Duration(timeout))
}
