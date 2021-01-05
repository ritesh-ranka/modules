package make

import (
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/persistenceOne/persistenceSDK/modules/orders/internal/module"
	"github.com/persistenceOne/persistenceSDK/schema/types/base"
	"github.com/persistenceOne/persistenceSDK/utilities/transaction"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_Make_Message(t *testing.T) {

	testFromID := base.NewID("fromID")
	testClassificationID := base.NewID("classificationID")
	testMakerOwnableID := base.NewID("makerOwnableID")
	testTakerOwnableID := base.NewID("takerOwnableID")
	testExpiresIn := base.NewHeight(12)
	testMakerOwnableSplit := sdkTypes.NewDec(2)

	fromAddress := "cosmos1pkkayn066msg6kn33wnl5srhdt3tnu2vzasz9c"
	fromAccAddress, Error := sdkTypes.AccAddressFromBech32(fromAddress)
	require.Nil(t, Error)

	immutableMetaTraits, Error := base.ReadMetaProperties("defaultImmutableMeta1:S|defaultImmutableMeta1")
	require.Equal(t, nil, Error)
	immutableTraits, Error := base.ReadProperties("defaultImmutable1:S|defaultImmutable1")
	require.Equal(t, nil, Error)
	mutableMetaTraits, Error := base.ReadMetaProperties("defaultMutableMeta1:S|defaultMutableMeta1")
	require.Equal(t, nil, Error)
	mutableTraits, Error := base.ReadProperties("defaultMutable1:S|defaultMutable1")
	require.Equal(t, nil, Error)

	testMessage := newMessage(fromAccAddress, testFromID, testClassificationID, testMakerOwnableID, testTakerOwnableID, testExpiresIn, testMakerOwnableSplit, immutableMetaTraits, immutableTraits, mutableMetaTraits, mutableTraits)
	require.Equal(t, message{From: fromAccAddress, FromID: testFromID, ClassificationID: testClassificationID, MakerOwnableID: testMakerOwnableID, TakerOwnableID: testTakerOwnableID, ExpiresIn: testExpiresIn, MakerOwnableSplit: testMakerOwnableSplit, ImmutableMetaProperties: immutableMetaTraits, ImmutableProperties: immutableTraits, MutableMetaProperties: mutableMetaTraits, MutableProperties: mutableTraits}, testMessage)
	require.Equal(t, module.Name, testMessage.Route())
	require.Equal(t, Transaction.GetName(), testMessage.Type())
	require.Equal(t, nil, testMessage.ValidateBasic())
	require.NotNil(t, message{}.ValidateBasic())
	require.Equal(t, sdkTypes.MustSortJSON(transaction.RegisterCodec(messagePrototype).MustMarshalJSON(testMessage)), testMessage.GetSignBytes())
	require.Equal(t, []sdkTypes.AccAddress{fromAccAddress}, testMessage.GetSigners())
	require.Equal(t, testMessage, messageFromInterface(testMessage))
	require.Equal(t, message{}, messageFromInterface(nil))
	require.Equal(t, message{}, messagePrototype())

}