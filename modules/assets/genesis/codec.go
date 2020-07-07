package genesis

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/persistenceOne/persistenceSDK/modules/assets/constants"
)

func registerCodec(codec *codec.Codec) {
	codec.RegisterConcrete(genesisState{}, fmt.Sprintf("/%v/%v", constants.ModuleName, "genesisState"), nil)
}

var packageCodec = codec.New()

func init() {
	registerCodec(packageCodec)
	packageCodec.Seal()
}