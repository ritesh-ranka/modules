package types

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	sdkTypesModule "github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type Module interface {
	sdkTypesModule.AppModuleBasic
	sdkTypesModule.AppModule

	GetStoreKey() string
	GetDefaultParamspace() string

	InitializeKeepers(*codec.Codec, sdkTypes.StoreKey, params.Subspace)
}
type module struct {
	ModuleName        string
	StoreKey          string
	DefaultParamspace string
	QueryRoute        string
	TransactionRoute  string
	GenesisState      GenesisState
	QueryList         []Query
	TransactionList   []Transaction
	Codec             func(*codec.Codec)
	Mapper            func(*codec.Codec, sdkTypes.StoreKey) Mapper
}

var _ Module = (*module)(nil)

func (module module) Name() string {
	return module.ModuleName
}
func (module module) RegisterCodec(codec *codec.Codec) {
	module.Codec(codec)
}
func (module module) DefaultGenesis(jsonMarshaler codec.JSONMarshaler) json.RawMessage {
	return jsonMarshaler.MustMarshalJSON(module.GenesisState.Default())
}
func (module module) ValidateGenesis(_ codec.JSONMarshaler, rawMessage json.RawMessage) error {
	module.GenesisState.Unmarshall(rawMessage)
	return module.GenesisState.Validate()
}
func (module module) RegisterRESTRoutes(cliContext context.CLIContext, router *mux.Router) {
	for _, query := range module.QueryList {
		router.HandleFunc(fmt.Sprintf("/%v/%v/{%v}", query.GetModuleName(), query.GetName(), "id"), query.RESTQueryHandler(cliContext)).Methods("GET")
	}

	for _, transaction := range module.TransactionList {
		router.HandleFunc(fmt.Sprintf("/%v/%v", transaction.GetModuleName(), transaction.GetName()), transaction.RESTRequestHandler(cliContext)).Methods("POST")
	}
}
func (module module) GetTxCmd(codec *codec.Codec) *cobra.Command {
	rootTransactionCommand := &cobra.Command{
		Use:                        module.TransactionRoute,
		Short:                      "Get root transaction command.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	var commandList []*cobra.Command
	for _, transaction := range module.TransactionList {
		commandList = append(commandList, transaction.Command(codec))
	}
	rootTransactionCommand.AddCommand(
		commandList...,
	)
	return rootTransactionCommand
}
func (module module) GetQueryCmd(codec *codec.Codec) *cobra.Command {
	rootQueryCommand := &cobra.Command{
		Use:                        module.QueryRoute,
		Short:                      "Get root query command.",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	var commandList []*cobra.Command
	for _, query := range module.QueryList {
		commandList = append(commandList, query.Command(codec))
	}
	rootQueryCommand.AddCommand(
		commandList...,
	)
	return rootQueryCommand
}
func (module module) RegisterInvariants(_ sdkTypes.InvariantRegistry) {}
func (module module) Route() string {
	return module.TransactionRoute
}
func (module module) NewHandler() sdkTypes.Handler {
	return func(context sdkTypes.Context, msg sdkTypes.Msg) (*sdkTypes.Result, error) {
		context = context.WithEventManager(sdkTypes.NewEventManager())

		for _, transaction := range module.TransactionList {
			if msg.Type() == transaction.GetName() {
				return transaction.HandleMessage(context, msg)
			}
		}
		return nil, errors.New(fmt.Sprintf("Unknown message type, %v for module %v", msg.Type(), module.Name()))
	}
}
func (module module) QuerierRoute() string {
	return module.QueryRoute
}
func (module module) NewQuerierHandler() sdkTypes.Querier {
	return func(context sdkTypes.Context, path []string, requestQuery abciTypes.RequestQuery) ([]byte, error) {
		for _, query := range module.QueryList {
			if query.GetName() == path[0] {
				return query.HandleMessage(context, requestQuery)
			}
		}
		return nil, errors.New(fmt.Sprintf("Unknown query path, %v for module %v", path[0], module.Name()))
	}
}
func (module module) InitGenesis(context sdkTypes.Context, _ codec.JSONMarshaler, rawMessage json.RawMessage) []abciTypes.ValidatorUpdate {
	module.GenesisState.Unmarshall(rawMessage)
	module.GenesisState.Initialize(context)
	return []abciTypes.ValidatorUpdate{}
}
func (module module) ExportGenesis(context sdkTypes.Context, _ codec.JSONMarshaler) json.RawMessage {
	return module.GenesisState.Export(context).Marshall()
}
func (module module) BeginBlock(_ sdkTypes.Context, _ abciTypes.RequestBeginBlock) {}

func (module module) EndBlock(_ sdkTypes.Context, _ abciTypes.RequestEndBlock) []abciTypes.ValidatorUpdate {
	return []abciTypes.ValidatorUpdate{}
}
func (module module) GetStoreKey() string {
	return module.StoreKey
}
func (module module) GetDefaultParamspace() string {
	return module.DefaultParamspace
}
func (module module) InitializeKeepers(codec *codec.Codec, storeKey sdkTypes.StoreKey, _ params.Subspace) {
	mapper := module.Mapper(codec, storeKey)

	for _, transaction := range module.TransactionList {
		transaction.InitializeKeeper(mapper)
	}

	for _, query := range module.QueryList {
		query.InitializeKeeper(mapper)
	}

	return
}
func NewModule(moduleName string, storeKey string, defaultParamspace string, queryRoute string, transactionRoute string, genesisState GenesisState, queryList []Query, transactionList []Transaction, codec func(*codec.Codec), mapper func(*codec.Codec, sdkTypes.StoreKey) Mapper) Module {
	return module{
		ModuleName:        moduleName,
		StoreKey:          storeKey,
		DefaultParamspace: defaultParamspace,
		QueryRoute:        queryRoute,
		TransactionRoute:  transactionRoute,
		GenesisState:      genesisState,
		QueryList:         queryList,
		TransactionList:   transactionList,
		Codec:             codec,
		Mapper:            mapper,
	}
}