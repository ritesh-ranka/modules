// Copyright [2021] - [2022], AssetMantle Pte. Ltd. and the code contributors
// SPDX-License-Identifier: Apache-2.0

package base

import (
	"strconv"

	"github.com/AssetMantle/modules/constants/errors"
	idsConstants "github.com/AssetMantle/modules/constants/ids"
	"github.com/AssetMantle/modules/schema/data"
	"github.com/AssetMantle/modules/schema/ids"
	baseIDs "github.com/AssetMantle/modules/schema/ids/base"
	"github.com/AssetMantle/modules/schema/traits"
	"github.com/AssetMantle/modules/schema/types"
)

type booleanData struct {
	Value bool `json:"value"`
}

var _ data.BooleanData = (*booleanData)(nil)

func (booleanData booleanData) GetID() ids.DataID {
	return baseIDs.NewDataID(booleanData)
}
func (booleanData booleanData) Compare(listable traits.Listable) int {
	compareBooleanData, Error := booleanDataFromInterface(listable)
	if Error != nil {
		panic(Error)
	}

	if booleanData.Value == compareBooleanData.Value {
		return 0
	} else if booleanData.Value == true {
		return 1
	}

	return -1
}
func (booleanData booleanData) String() string {
	return strconv.FormatBool(booleanData.Value)
}
func (booleanData booleanData) GetType() types.ID {
	return idsConstants.BooleanDataID
}
func (booleanData booleanData) ZeroValue() types.Data {
	return NewBooleanData(false)
}
func (booleanData booleanData) GenerateHash() types.ID {
	if booleanData.Compare(booleanData.ZeroValue()) == 0 {
		return baseIDs.NewID(strconv.FormatBool(false))
	}

	return baseIDs.NewID(strconv.FormatBool(true))
}
func (booleanData booleanData) Get() bool {
	return booleanData.Value
}

func booleanDataFromInterface(listable traits.Listable) (booleanData, error) {
	switch value := listable.(type) {
	case booleanData:
		return value, nil
	default:
		return booleanData{}, errors.MetaDataError
	}
}

func NewBooleanData(value bool) types.Data {
	return booleanData{
		Value: value,
	}
}

func ReadBooleanData(dataString string) (types.Data, error) {
	if dataString == "" {
		return booleanData{}.ZeroValue(), nil
	}

	Bool, Error := strconv.ParseBool(dataString)
	if Error != nil {
		return booleanData{}.ZeroValue(), Error
	}

	return NewBooleanData(Bool), nil
}