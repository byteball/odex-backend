package endpoints

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
	"github.com/byteball/odex-backend/utils/testutils/mocks"
	"github.com/gorilla/mux"
)

func SetupTradeTest() (*mux.Router, *mocks.TradeService) {
	r := mux.NewRouter()
	tradeService := new(mocks.TradeService)

	ServeTradeResource(r, tradeService)

	return r, tradeService
}

func TestHandleGetTradeHistory(t *testing.T) {
	router, tradeService := SetupTradeTest()

	t1 := testutils.GetTestZRXToken()
	t2 := testutils.GetTestWETHToken()

	tr1 := types.Trade{}
	tr2 := types.Trade{}
	trs := []*types.Trade{&tr1, &tr2}

	tradeService.On("GetSortedTrades", t1.Asset, t2.Asset, 20).Return(trs, nil)

	req, err := http.NewRequest("GET", "/trades/pair?baseToken="+url.QueryEscape(t1.Asset)+"&quoteToken="+url.QueryEscape(t2.Asset), nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusOK)
	}

	json.NewDecoder(rr.Body)

}
