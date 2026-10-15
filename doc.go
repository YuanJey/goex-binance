// Package binance implements a Binance USDⓈ-M Futures client.
//
// Prices, quantities, fees, and PnL are represented as decimal strings so no
// precision is lost at the API boundary. The Client exposes native Binance
// semantics through Market, Account, Trade, Algo, and Stream services.
package binance
