// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Exchange   string `config:"exchange"`
	SymbolFrom string `config:"symbol_from"`
	SymbolTo   string `config:"symbol_to"`
}

var DefaultConfig = Config{
	Exchange:   "",
	SymbolFrom: "",
	SymbolTo:   "USD",
}
