// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	Type       int    `config:"type"`
	Exchange   string `config:"exchange"`
	SymbolFrom string `config:"symbol_from"`
	SymbolTo   string `config:"symbol_to"`
}

var DefaultConfig = Config{
	Type:       0,
	Exchange:   "",
	SymbolFrom: "",
	SymbolTo:   "USD",
}
