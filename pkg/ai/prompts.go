/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
package ai

const (
	DefaultPrompt = `Simplify the following Kubernetes error message delimited by triple dashes written in English language; --- %s ---.
	Provide the most possible solution in a step by step style in no more than 560 characters. Write the output in the following format:
	Error: {Explain error here}
	Solution: {Step by step solution here}
	`
	KnowledgePrompt = "In kubernetes, what can cause %s?"
)
