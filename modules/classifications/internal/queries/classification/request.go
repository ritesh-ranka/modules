/*
 Copyright [2019] - [2021], PERSISTENCE TECHNOLOGIES PTE. LTD. and the persistenceSDK contributors
 SPDX-License-Identifier: Apache-2.0
*/

package classification

import (
	"github.com/asaskevich/govalidator"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/persistenceOne/persistenceSDK/constants/flags"
	"github.com/persistenceOne/persistenceSDK/modules/classifications/internal/common"
	"github.com/persistenceOne/persistenceSDK/schema/helpers"
	"github.com/persistenceOne/persistenceSDK/schema/types"
	"github.com/persistenceOne/persistenceSDK/schema/types/base"
)

type queryRequest struct {
	ClassificationID types.ID `json:"classificationID" valid:"required~required field classificationID missing"`
}

var _ helpers.QueryRequest = (*queryRequest)(nil)

// Validate godoc
// @Summary Search for an identity by identity ID
// @Description Able to query the asset
// @Accept json
// @Produce json
// @Tags Classifications
// @Param classificationID path string true "Unique identifier of an asset classification."
// @Success 200 {object} queryResponse "Message for a successful search response."
// @Failure default  {object}  queryResponse "Message for an unexpected error response."
// @Router /classifications/classifications/{classificationID} [get]
func (queryRequest queryRequest) Validate() error {
	_, Error := govalidator.ValidateStruct(queryRequest)
	return Error
}
func (queryRequest queryRequest) FromCLI(cliCommand helpers.CLICommand, _ context.CLIContext) helpers.QueryRequest {
	return newQueryRequest(base.NewID(cliCommand.ReadString(flags.ClassificationID)))
}

func (queryRequest queryRequest) FromMap(vars map[string]string) helpers.QueryRequest {
	return newQueryRequest(base.NewID(vars[Query.GetName()]))
}
func (queryRequest queryRequest) Encode() ([]byte, error) {
	return common.Codec.MarshalJSON(queryRequest)
}
func (queryRequest queryRequest) Decode(bytes []byte) (helpers.QueryRequest, error) {
	if Error := common.Codec.UnmarshalJSON(bytes, &queryRequest); Error != nil {
		return nil, Error
	}

	return queryRequest, nil
}
func requestPrototype() helpers.QueryRequest {
	return queryRequest{}
}
func queryRequestFromInterface(request helpers.QueryRequest) queryRequest {
	switch value := request.(type) {
	case queryRequest:
		return value
	default:
		return queryRequest{}
	}
}
func newQueryRequest(classificationID types.ID) helpers.QueryRequest {
	return queryRequest{ClassificationID: classificationID}
}
