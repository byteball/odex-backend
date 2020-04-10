package endpoints

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/byteball/odex-backend/services"
	"github.com/byteball/odex-backend/types"
	"github.com/byteball/odex-backend/utils/testutils"
	"github.com/byteball/odex-backend/utils/testutils/mocks"
	"github.com/gorilla/mux"
)

func SetupPairEndpointTest() (*mux.Router, *mocks.PairService) {
	r := mux.NewRouter()
	pairService := new(mocks.PairService)
	tokenService := new(mocks.TokenService)

	ServePairResource(r, pairService, tokenService)

	return r, pairService
}

func TestHandleCreatePair(t *testing.T) {
	router, pairService := SetupPairEndpointTest()

	pair := types.Pair{
		BaseTokenSymbol:  "ZRX",
		BaseAsset:        "0x1",
		QuoteTokenSymbol: "WETH",
		QuoteAsset:       "0x2",
	}

	pairService.On("Create", &pair).Return(nil)

	b, _ := json.Marshal(pair)
	req, err := http.NewRequest("POST", "/pair/create", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusCreated)
	}

	created := struct{ Data types.Pair }{}
	json.NewDecoder(rr.Body).Decode(&created)

	pairService.AssertCalled(t, "Create", &pair)
	testutils.ComparePair(t, &pair, &created.Data)
}

func TestHandleCreateInvalidPair(t *testing.T) {
	router, pairService := SetupPairEndpointTest()

	pair := types.Pair{
		BaseTokenSymbol: "ZRX",
		BaseAsset:       "0x1",
		QuoteAsset:      "0x2",
	}

	pairService.On("Create", &pair).Return(services.ErrBaseTokenNotFound)

	b, _ := json.Marshal(pair)
	req, err := http.NewRequest("POST", "/pair/create", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusBadRequest)
	}
}

func TestHandleGetAllPairs(t *testing.T) {
	router, pairService := SetupPairEndpointTest()

	p1 := types.Pair{
		BaseTokenSymbol: "ZRX",
		BaseAsset:       "0x1",
		QuoteAsset:      "0x2",
	}

	p2 := types.Pair{
		BaseTokenSymbol: "WETH",
		BaseAsset:       "0x3",
		QuoteAsset:      "0x4",
	}

	pairService.On("GetAll").Return([]types.Pair{p1, p2}, nil)

	req, err := http.NewRequest("GET", "/pairs", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusOK)
	}

	result := struct{ Data []types.Pair }{}
	json.NewDecoder(rr.Body).Decode(&result)

	pairService.AssertCalled(t, "GetAll")
	testutils.ComparePair(t, &p1, &result.Data[0])
	testutils.ComparePair(t, &p2, &result.Data[1])
}

func TestHandleGetPair(t *testing.T) {
	router, pairService := SetupPairEndpointTest()

	base := "PP7/+yQc6+XAZ1WBsUzGwcmfIlInIRLsLlfoWJc/3kY="
	quote := "7l7GzugRUz9b/q7M+A1K9IIQ5yWnqlB6CyImXx73TQs="

	p1 := types.Pair{
		BaseTokenSymbol:  "ZRX",
		QuoteTokenSymbol: "WETH",
		BaseAsset:        base,
		QuoteAsset:       quote,
	}

	pairService.On("GetByAsset", base, quote).Return(&p1, nil)

	url := "/pair?baseToken=" + url.QueryEscape(base) + "&quoteToken=" + url.QueryEscape(quote)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusOK)
	}

	result := struct{ Data types.Pair }{}
	json.NewDecoder(rr.Body).Decode(&result)

	pairService.AssertCalled(t, "GetByAsset", base, quote)
	testutils.ComparePair(t, &p1, &result.Data)
}
