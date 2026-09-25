// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package lint

// Lint thresholds for semantic checks.
//
//nolint:revive // intentionally exported for use by consumers
const (
	ProvenanceDriftThreshold = 0.20
	FreshnessHalfLifeDays    = 30
	FreshnessScoreThreshold  = 50.0
	SummaryMinLength         = 10
	SummaryMaxLength         = 200
)
