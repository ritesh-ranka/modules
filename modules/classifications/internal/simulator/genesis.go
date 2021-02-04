/*
 Copyright [2019] - [2021], PERSISTENCE TECHNOLOGIES PTE. LTD. and the persistenceSDK contributors
 SPDX-License-Identifier: Apache-2.0
*/

package simulator

import (
	"math/rand"

	"github.com/persistenceOne/persistenceSDK/modules/classifications/internal/common"

	"github.com/persistenceOne/persistenceSDK/modules/classifications/internal/parameters"
	"github.com/persistenceOne/persistenceSDK/schema/helpers"
	baseSimulation "github.com/persistenceOne/persistenceSDK/utilities/simulation/schema/types/base"

	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/persistenceOne/persistenceSDK/modules/classifications/internal/key"
	"github.com/persistenceOne/persistenceSDK/modules/classifications/internal/mappable"
	classificationsModule "github.com/persistenceOne/persistenceSDK/modules/classifications/internal/module"
	"github.com/persistenceOne/persistenceSDK/modules/classifications/internal/parameters/dummy"
	baseHelpers "github.com/persistenceOne/persistenceSDK/schema/helpers/base"
	"github.com/persistenceOne/persistenceSDK/schema/types"
	"github.com/persistenceOne/persistenceSDK/schema/types/base"
)

func (simulator) RandomizedGenesisState(simulationState *module.SimulationState) {
	var data types.Data

	simulationState.AppParams.GetOrGenerate(
		simulationState.Cdc,
		dummy.ID.String(),
		&data,
		simulationState.Rand,
		func(rand *rand.Rand) { data = base.NewDecData(sdkTypes.NewDecWithPrec(int64(rand.Intn(99)), 2)) },
	)

	mappableList := make([]helpers.Mappable, simulationState.Rand.Intn(99))

	for i := range mappableList {
		immutables := baseSimulation.GenerateRandomImmutables(simulationState.Rand)
		mutables := baseSimulation.GenerateRandomMutables(simulationState.Rand)
		mappableList[i] = mappable.NewClassification(key.NewClassificationID(baseSimulation.GenerateRandomID(simulationState.Rand), immutables, mutables), immutables, mutables)
	}

	genesisState := baseHelpers.NewGenesis(key.Prototype, mappable.Prototype, nil, parameters.Prototype().GetList()).Initialize(mappableList, []types.Parameter{dummy.Parameter.Mutate(data)})

	simulationState.GenState[classificationsModule.Name] = common.Codec.MustMarshalJSON(genesisState)
}
