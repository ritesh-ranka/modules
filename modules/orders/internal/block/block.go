/*
 Copyright [2019] - [2021], PERSISTENCE TECHNOLOGIES PTE. LTD. and the persistenceSDK contributors
 SPDX-License-Identifier: Apache-2.0
*/

package block

import (
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/persistenceOne/persistenceSDK/constants/errors"
	"github.com/persistenceOne/persistenceSDK/constants/properties"
	"github.com/persistenceOne/persistenceSDK/modules/metas/auxiliaries/scrub"
	"github.com/persistenceOne/persistenceSDK/modules/metas/auxiliaries/supplement"
	"github.com/persistenceOne/persistenceSDK/modules/orders/internal/key"
	"github.com/persistenceOne/persistenceSDK/modules/orders/internal/mappable"
	"github.com/persistenceOne/persistenceSDK/modules/orders/internal/module"
	"github.com/persistenceOne/persistenceSDK/modules/splits/auxiliaries/transfer"
	"github.com/persistenceOne/persistenceSDK/schema/helpers"
	"github.com/persistenceOne/persistenceSDK/schema/mappables"
	"github.com/persistenceOne/persistenceSDK/schema/types"
	"github.com/persistenceOne/persistenceSDK/schema/types/base"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type block struct {
	mapper              helpers.Mapper
	parameters          helpers.Parameters
	supplementAuxiliary helpers.Auxiliary
	transferAuxiliary   helpers.Auxiliary
	scrubAuxiliary      helpers.Auxiliary
}

var _ helpers.Block = (*block)(nil)

func (block block) Begin(_ sdkTypes.Context, _ abciTypes.RequestBeginBlock) {

}

func (block block) End(context sdkTypes.Context, _ abciTypes.RequestEndBlock) {
	executeOrders := make(map[types.ID]bool)
	orders := block.mapper.NewCollection(context)

	orders.Iterate(
		key.FromID(base.NewID("")),
		func(order helpers.Mappable) bool {
			metaProperties, err := supplement.GetMetaPropertiesFromResponse(block.supplementAuxiliary.GetKeeper().Help(context, supplement.NewAuxiliaryRequest(order.(mappables.Order).GetExpiry(), order.(mappables.Order).GetMakerOwnableSplit())))
			if err != nil {
				panic(err)
			}
			if expiryProperty := metaProperties.Get(base.NewID(properties.Expiry)); expiryProperty != nil {
				expiry, err := expiryProperty.GetMetaFact().GetData().AsHeight()
				if err != nil {
					panic(err)
				}

				if expiry.Compare(base.NewHeight(context.BlockHeight())) <= 0 {
					makerOwnableSplitProperty := metaProperties.Get(base.NewID(properties.MakerOwnableSplit))
					if makerOwnableSplitProperty == nil {
						panic(errors.MetaDataError)
					}
					makerOwnableSplit, err := makerOwnableSplitProperty.GetMetaFact().GetData().AsDec()
					if err != nil {
						panic(err)
					}
					if auxiliaryResponse := block.transferAuxiliary.GetKeeper().Help(context, transfer.NewAuxiliaryRequest(base.NewID(module.Name), order.(mappables.Order).GetMakerID(), order.(mappables.Order).GetMakerOwnableID(), makerOwnableSplit)); !auxiliaryResponse.IsSuccessful() {
						panic(auxiliaryResponse.GetError())
					}
					orders.Remove(order)
				} else {
					id1 := key.NewOrderID(order.(mappables.Order).GetClassificationID(), order.(mappables.Order).GetMakerOwnableID(), order.(mappables.Order).GetTakerOwnableID(), base.NewID(""), base.NewID(""), base.NewID(""), base.NewProperties())
					id2 := key.NewOrderID(order.(mappables.Order).GetClassificationID(), order.(mappables.Order).GetTakerOwnableID(), order.(mappables.Order).GetMakerOwnableID(), base.NewID(""), base.NewID(""), base.NewID(""), base.NewProperties())
					if !executeOrders[id1] && !executeOrders[id2] {
						executeOrders[id1] = true
					}
				}
			}
			return false
		},
	)

	for partialOrderID := range executeOrders {
		nextPartialOrderID := false

		orders.Iterate(key.FromID(partialOrderID), func(orderMappable helpers.Mappable) bool {
			orders.Iterate(
				key.FromID(key.NewOrderID(orderMappable.(mappables.Order).GetClassificationID(), orderMappable.(mappables.Order).GetTakerOwnableID(), orderMappable.(mappables.Order).GetMakerOwnableID(), base.NewID(""), base.NewID(""), base.NewID(""), base.NewProperties())),
				func(executableMappableOrder helpers.Mappable) bool {

					var leftOrder mappables.Order
					var rightOrder mappables.Order

					orderHeight, err := orderMappable.(mappables.Order).GetCreation().GetMetaFact().GetData().AsHeight()
					if err != nil {
						panic(err)
					}

					executableOrderHeight, err := executableMappableOrder.(mappables.Order).GetCreation().GetMetaFact().GetData().AsHeight()
					if err != nil {
						panic(err)
					}

					switch {
					case orderHeight.Compare(executableOrderHeight) > 0:
						leftOrder = orderMappable.(mappables.Order)
						rightOrder = executableMappableOrder.(mappables.Order)
					case executableOrderHeight.Compare(orderHeight) > 0:
						leftOrder = executableMappableOrder.(mappables.Order)
						rightOrder = orderMappable.(mappables.Order)
					default:
						// TODO
						leftOrder = orderMappable.(mappables.Order)
						rightOrder = executableMappableOrder.(mappables.Order)
					}

					leftOrderExchangeRate, err := leftOrder.GetExchangeRate().GetMetaFact().GetData().AsDec()
					if err != nil {
						panic(err)
					}

					leftOrderMetaProperties, err := supplement.GetMetaPropertiesFromResponse(block.supplementAuxiliary.GetKeeper().Help(context, supplement.NewAuxiliaryRequest(leftOrder.GetMakerOwnableSplit())))
					if err != nil {
						panic(err)
					}

					leftOrderMakerOwnableSplit, err := leftOrderMetaProperties.Get(base.NewID(properties.MakerOwnableSplit)).GetMetaFact().GetData().AsDec()
					if err != nil {
						panic(err)
					}

					rightOrderExchangeRate, err := rightOrder.GetExchangeRate().GetMetaFact().GetData().AsDec()
					if err != nil {
						panic(err)
					}

					rightOrderMetaProperties, err := supplement.GetMetaPropertiesFromResponse(block.supplementAuxiliary.GetKeeper().Help(context, supplement.NewAuxiliaryRequest(rightOrder.GetMakerOwnableSplit())))
					if err != nil {
						panic(err)
					}

					rightOrderMakerOwnableSplit, err := rightOrderMetaProperties.Get(base.NewID(properties.MakerOwnableSplit)).GetMetaFact().GetData().AsDec()
					if err != nil {
						panic(err)
					}

					rightOrderTakerOwnableSplitDemanded := rightOrderExchangeRate.MulTruncate(rightOrderMakerOwnableSplit).MulTruncate(sdkTypes.SmallestDec())

					if leftOrderExchangeRate.MulTruncate(rightOrderExchangeRate).MulTruncate(sdkTypes.SmallestDec()).MulTruncate(sdkTypes.SmallestDec()).LTE(sdkTypes.OneDec()) {
						switch {
						case leftOrderMakerOwnableSplit.GT(rightOrderTakerOwnableSplitDemanded):
							if auxiliaryResponse := block.transferAuxiliary.GetKeeper().Help(context, transfer.NewAuxiliaryRequest(base.NewID(module.Name), leftOrder.GetMakerID(), leftOrder.GetTakerOwnableID(), rightOrderMakerOwnableSplit)); !auxiliaryResponse.IsSuccessful() {
								panic(auxiliaryResponse.GetError())
							}
							if auxiliaryResponse := block.transferAuxiliary.GetKeeper().Help(context, transfer.NewAuxiliaryRequest(base.NewID(module.Name), rightOrder.GetMakerID(), leftOrder.GetMakerOwnableID(), rightOrderTakerOwnableSplitDemanded)); !auxiliaryResponse.IsSuccessful() {
								panic(auxiliaryResponse.GetError())
							}

							mutableProperties, err := scrub.GetPropertiesFromResponse(block.scrubAuxiliary.GetKeeper().Help(context, scrub.NewAuxiliaryRequest(base.NewMetaProperty(base.NewID(properties.MakerOwnableSplit), base.NewMetaFact(base.NewDecData(leftOrderMakerOwnableSplit.Sub(rightOrderTakerOwnableSplitDemanded)))))))
							if err != nil {
								panic(err)
							}

							orders.Mutate(mappable.NewOrder(leftOrder.GetID(), leftOrder.GetImmutableProperties(), leftOrder.GetMutableProperties().Mutate(mutableProperties.GetList()...)))
							orders.Remove(rightOrder)

							if executableOrderHeight.Compare(orderHeight) > 0 {
								return true
							}
						case leftOrderMakerOwnableSplit.LT(rightOrderTakerOwnableSplitDemanded):
							sendToLeftOrder := leftOrderMakerOwnableSplit.QuoTruncate(sdkTypes.SmallestDec()).QuoTruncate(rightOrderExchangeRate)
							if auxiliaryResponse := block.transferAuxiliary.GetKeeper().Help(context, transfer.NewAuxiliaryRequest(base.NewID(module.Name), leftOrder.GetMakerID(), leftOrder.GetTakerOwnableID(), sendToLeftOrder)); !auxiliaryResponse.IsSuccessful() {
								panic(auxiliaryResponse.GetError())
							}
							if auxiliaryResponse := block.transferAuxiliary.GetKeeper().Help(context, transfer.NewAuxiliaryRequest(base.NewID(module.Name), rightOrder.GetMakerID(), leftOrder.GetMakerOwnableID(), leftOrderMakerOwnableSplit)); !auxiliaryResponse.IsSuccessful() {
								panic(auxiliaryResponse.GetError())
							}

							mutableProperties, err := scrub.GetPropertiesFromResponse(block.scrubAuxiliary.GetKeeper().Help(context, scrub.NewAuxiliaryRequest(base.NewMetaProperty(base.NewID(properties.MakerOwnableSplit), base.NewMetaFact(base.NewDecData(rightOrderMakerOwnableSplit.Sub(sendToLeftOrder)))))))
							if err != nil {
								panic(err)
							}

							orders.Mutate(mappable.NewOrder(rightOrder.GetID(), rightOrder.GetImmutableProperties(), rightOrder.GetMutableProperties().Mutate(mutableProperties.GetList()...)))
							orders.Remove(leftOrder)

							if orderHeight.Compare(executableOrderHeight) >= 0 {
								return true
							}
						default:
							// case leftOrderMakerOwnableSplit.Equal(rightOrderTakerOwnableSplitDemanded):
							if auxiliaryResponse := block.transferAuxiliary.GetKeeper().Help(context, transfer.NewAuxiliaryRequest(base.NewID(module.Name), leftOrder.GetMakerID(), leftOrder.GetTakerOwnableID(), rightOrderMakerOwnableSplit)); !auxiliaryResponse.IsSuccessful() {
								panic(auxiliaryResponse.GetError())
							}
							if auxiliaryResponse := block.transferAuxiliary.GetKeeper().Help(context, transfer.NewAuxiliaryRequest(base.NewID(module.Name), rightOrder.GetMakerID(), leftOrder.GetMakerOwnableID(), leftOrderMakerOwnableSplit)); !auxiliaryResponse.IsSuccessful() {
								panic(auxiliaryResponse.GetError())
							}

							orders.Remove(rightOrder)
							orders.Remove(leftOrder)
							return true
						}
					} else {
						nextPartialOrderID = true
						return true
					}
					return false
				},
			)
			return nextPartialOrderID
		})

		if nextPartialOrderID {
			continue
		}
	}
}

func (block block) Initialize(mapper helpers.Mapper, parameters helpers.Parameters, auxiliaryKeepers ...interface{}) helpers.Block {
	block.mapper, block.parameters = mapper, parameters

	for _, auxiliaryKeeper := range auxiliaryKeepers {
		switch value := auxiliaryKeeper.(type) {
		case helpers.Auxiliary:
			switch value.GetName() {
			case supplement.Auxiliary.GetName():
				block.supplementAuxiliary = value
			case transfer.Auxiliary.GetName():
				block.transferAuxiliary = value
			case scrub.Auxiliary.GetName():
				block.scrubAuxiliary = value
			}
		default:
			panic(errors.UninitializedUsage)
		}
	}

	return block
}
