package endpoints

import (
	"bytes"
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

func SetupTokenTest() (*mux.Router, *mocks.TokenService) {
	r := mux.NewRouter()
	tokenService := new(mocks.TokenService)

	ServeTokenResource(r, tokenService)

	return r, tokenService
}

func TestHandleCreateTokens(t *testing.T) {
	router, tokenService := SetupTokenTest()

	token := types.Token{
		//Name:    "ZRX",
		Symbol:   "ZRX",
		Decimals: 18,
		Quote:    false,
		Asset:    "0x1",
	}

	tokenService.On("Create", &token).Return(nil)

	b, _ := json.Marshal(token)
	req, err := http.NewRequest("POST", "/tokens", bytes.NewBuffer(b))
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusCreated)
	}

	created := struct {
		Data types.Token
	}{}
	json.NewDecoder(rr.Body).Decode(&created)

	tokenService.AssertCalled(t, "Create", &token)
	testutils.CompareToken(t, &token, &created.Data)
}

func TestHandleGetTokens(t *testing.T) {
	router, tokenService := SetupTokenTest()

	t1 := testutils.GetTestZRXToken()
	t2 := testutils.GetTestWETHToken()

	tokenService.On("GetAll").Return([]types.Token{t1, t2}, nil)

	req, err := http.NewRequest("GET", "/tokens", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusOK)
	}

	result := struct {
		Data []types.Token
	}{}
	json.NewDecoder(rr.Body).Decode(&result)

	tokenService.AssertCalled(t, "GetAll")
	testutils.CompareToken(t, &t1, &result.Data[0])
	testutils.CompareToken(t, &t2, &result.Data[1])
}

func TestHandleGetQuoteTokens(t *testing.T) {
	router, tokenService := SetupTokenTest()

	t1 := types.Token{
		//Name:    "WETH",
		Symbol:   "WETH",
		Decimals: 18,
		Quote:    true,
		Asset:    "0x1",
	}

	t2 := types.Token{
		//Name:    "DAI",
		Symbol:   "DAI",
		Decimals: 18,
		Quote:    true,
		Asset:    "0x2",
	}

	tokenService.On("GetQuoteTokens").Return([]types.Token{t1, t2}, nil)

	req, err := http.NewRequest("GET", "/tokens/quote", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusOK)
	}

	result := struct {
		Data []types.Token
	}{}
	json.NewDecoder(rr.Body).Decode(&result)

	tokenService.AssertCalled(t, "GetQuoteTokens")
	testutils.CompareToken(t, &t1, &result.Data[0])
	testutils.CompareToken(t, &t2, &result.Data[1])
}

func TestHandleGetBaseTokens(t *testing.T) {
	router, tokenService := SetupTokenTest()

	t1 := types.Token{
		//Name:    "WETH",
		Symbol:   "WETH",
		Decimals: 18,
		Quote:    false,
		Asset:    "0x1",
	}

	t2 := types.Token{
		//Name:    "DAI",
		Symbol:   "DAI",
		Decimals: 18,
		Quote:    false,
		Asset:    "0x2",
	}

	tokenService.On("GetBaseTokens").Return([]types.Token{t1, t2}, nil)

	req, err := http.NewRequest("GET", "/tokens/base", nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusOK)
	}

	result := struct {
		Data []types.Token
	}{}
	json.NewDecoder(rr.Body).Decode(&result)

	tokenService.AssertCalled(t, "GetBaseTokens")
	testutils.CompareToken(t, &t1, &result.Data[0])
	testutils.CompareToken(t, &t2, &result.Data[1])
}

func TestHandleGetToken(t *testing.T) {
	router, tokenService := SetupTokenTest()

	asset := "xCI8oRWJFavpgx3Wi7w+9A/0hSaxg3iqkp7h1buMjGc="

	t1 := types.Token{
		//Name:    "DAI",
		Symbol:   "DAI",
		Decimals: 18,
		Quote:    false,
		Asset:    asset,
	}

	tokenService.On("GetByAssetOrSymbol", asset).Return(&t1, nil)

	url := "/tokens/" + url.PathEscape(asset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Handler return wrong status. Got %v want %v", rr.Code, http.StatusOK)
	}

	result := struct {
		Data types.Token
	}{}
	json.NewDecoder(rr.Body).Decode(&result)

	tokenService.AssertCalled(t, "GetByAssetOrSymbol", asset)
	testutils.Compare(t, &t1, &result.Data)
}

// func TestHandleGetTokens(t *testing.T) {
// 	router, tokenService := SetupTokenTest()

// }

// func TestHandleGetQuoteTokens(t *testing.T) {
// 	router, tokenService := SetupTokenTest()

// }

// func TestHandleGetBaseTokens(t *testing.T) {
// 	router, tokenService := SetupTokenTest()
// }

// func TestHandleGetToken(t *testing.T) {
// 	router, tokenService := SetupTokenTest()

// }

// var resp interface{}
// 			if err := json.Unmarshal(res.Body.Bytes(), &resp); err != nil {
// 				fmt.Printf("%v", err)
// 			}

// 			bytes.NewBufferString
