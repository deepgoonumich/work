package rstore

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.benzinga.io/benzinga/ftp-engine/config"
	"gitlab.benzinga.io/benzinga/reference-service/reference"
)

func TestTicker(t *testing.T) {
	// Load Config
	cfg, err := config.LoadConfig("test")
	require.NoError(t, err)

	logger, err := cfg.LoadLogger()
	require.NoError(t, err)

	c, err := NewClient(logger, cfg.RedisURL)
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, c.Status(ctx))

	var data reference.FinancialData

	require.NoError(t, json.Unmarshal(instrumentsJSON, &data))

	for _, v := range data.Instruments {
		instr := v
		assert.NoError(t, c.PutSymbolCurrency(ctx, &instr))
		assert.NoError(t, c.PutSymbolExchange(ctx, &instr))

		res, err := c.GetSymbolCurrency(ctx, instr.Symbol, instr.CurrencyID)
		require.NoError(t, err)
		assert.Equal(t, instr.ISIN, res.ISIN)
		assert.Equal(t, instr.Symbol, res.Symbol)
		assert.Equal(t, instr.CurrencyID, res.CurrencyID)

		res, err = c.GetSymbolExchange(ctx, instr.Symbol, instr.Exchange)
		require.NoError(t, err)
		assert.Equal(t, instr.ISIN, res.ISIN)
		assert.Equal(t, instr.Symbol, res.Symbol)
		assert.Equal(t, instr.Exchange, res.Exchange)
	}

}

var instrumentsJSON = []byte(`{
	"instruments": [{
			"symbol": "A",
			"currencyId": "CAD",
			"exchange": "TSX",
			"exchangeISO": "TSX",
			"type": "STOCK",
			"cik": "1463915",
			"cusip": "04226J108",
			"nameShort": "Armor Minerals",
			"ipoDate": "2008-01-29",
			"sharesOutstanding": 44319015,
			"assetClassification": {
				"morningstar": {
					"superSectorCode": 1,
					"superSectorName": "Cyclical",
					"sectorCode": 101,
					"sectorName": "Basic Materials",
					"groupCode": 10106,
					"groupName": "Metals \u0026 Mining",
					"industryCode": 10106011,
					"industryName": "Industrial Metals \u0026 Minerals"
				},
				"sic": {
					"code": 619,
					"name": "Other Metal Mines"
				},
				"naics": {
					"code": 212299,
					"sectorCode": 21,
					"sectorName": "Mining, Quarrying, and Oil and Gas Extraction",
					"subSectorCode": 212,
					"subSectorName": "Mining (except Oil and Gas)",
					"industryGroupCode": 2122,
					"industryGroupName": "Metal Ore Mining",
					"industryCode": 21229,
					"industryName": "Other Metal Ore Mining",
					"nationalIndustryCode": 212299,
					"nationalIndustryName": "All Other Metal Ore Mining "
				}
			}
		},
		{
			"symbol": "A",
			"currencyId": "USD",
			"exchange": "NYSE",
			"exchangeISO": "NYS",
			"type": "STOCK",
			"cik": "1090872",
			"cusip": "00846U101",
			"isin": "US00846U1016",
			"nameShort": "Agilent Technologies",
			"ipoDate": "1999-11-18",
			"sharesOutstanding": 315993352,
			"shareFloat": 316507619,
			"assetClassification": {
				"morningstar": {
					"superSectorCode": 2,
					"superSectorName": "Defensive",
					"sectorCode": 206,
					"sectorName": "Healthcare",
					"groupCode": 20640,
					"groupName": "Medical Diagnostics \u0026 Research",
					"industryCode": 20640091,
					"industryName": "Diagnostics \u0026 Research"
				},
				"sic": {
					"code": 3825,
					"name": "Instruments For Measuring \u0026 Testing Of Electricity"
				},
				"naics": {
					"code": 334516,
					"sectorCode": 33,
					"sectorName": "Manufacturing",
					"subSectorCode": 334,
					"subSectorName": "Computer and Electronic Product Manufacturing",
					"industryGroupCode": 3345,
					"industryGroupName": "Navigational, Measuring, Electromedical, and Control Instruments Manufacturing",
					"industryCode": 33451,
					"industryName": "Navigational, Measuring, Electromedical, and Control Instruments Manufacturing",
					"nationalIndustryCode": 334516,
					"nationalIndustryName": "Analytical Laboratory Instrument Manufacturing "
				}
			}
		}
	]
}`)
