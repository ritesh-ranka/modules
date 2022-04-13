// Copyright [2021] - [2022], AssetMantle Pte. Ltd. and the code contributors
// SPDX-License-Identifier: Apache-2.0

package ids

import (
	"github.com/AssetMantle/modules/schema/types"
)

type DataID interface {
	types.ID
	GetHash() types.ID
}