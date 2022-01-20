/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package service

import (
	"context"

	"github.com/fidelity/theliv/internal/problem"
)

type ContextKey int

const (
	DetecotrInputKey ContextKey = iota
)

func GetDetectorInput(ctx context.Context) *problem.DetectorCreationInput {
	return ctx.Value(DetecotrInputKey).(*problem.DetectorCreationInput)
}

func SetDetectorInput(ctx context.Context, input *problem.DetectorCreationInput) context.Context {
	return context.WithValue(ctx, DetecotrInputKey, input)
}
