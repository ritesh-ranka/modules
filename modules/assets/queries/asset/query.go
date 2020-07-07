package asset

import (
	"github.com/persistenceOne/persistenceSDK/modules/assets/constants"
	"github.com/persistenceOne/persistenceSDK/types"
)

var Query = types.NewQuery(
	constants.ModuleName,
	constants.AssetQuery,
	NewQueryKeeper,
	constants.AssetQueryShort,
	constants.AssetQueryLong,
	queryRequestPrototype,
	queryResponsePrototype,
	packageCodec,
	registerCodec,
	[]types.CLIFlag{constants.AssetID},
)