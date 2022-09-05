package handlers

import (
	"encoding/json"
	"eth2-exporter/db"
	"eth2-exporter/types"
	"eth2-exporter/utils"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
)

var eth1AddressTemplate = template.Must(template.New("address").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/address.html"))

func Eth1Address(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	addressBytes := common.FromHex(address)
	data := InitPageData(w, r, "address", "/address", "Address")

	metadata, err := db.BigtableClient.GetMetadataForAddress(common.FromHex(address))

	if err != nil {
		logger.Errorf("error retieving balances for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
	g := new(errgroup.Group)
	g.SetLimit(7)

	var txns *types.DataTableResponse
	var internal *types.DataTableResponse
	var erc20 *types.DataTableResponse
	var erc721 *types.DataTableResponse
	var erc1155 *types.DataTableResponse
	var blocksMined *types.DataTableResponse
	var unclesMined *types.DataTableResponse

	g.Go(func() error {
		var err error
		txns, err = db.BigtableClient.GetAddressTransactionsTableData(addressBytes, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		internal, err = db.BigtableClient.GetAddressInternalTableData(addressBytes, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc20, err = db.BigtableClient.GetAddressErc20TableData(addressBytes, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc721, err = db.BigtableClient.GetAddressErc721TableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		erc1155, err = db.BigtableClient.GetAddressErc1155TableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		blocksMined, err = db.BigtableClient.GetAddressBlocksMinedTableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		unclesMined, err = db.BigtableClient.GetAddressUnclesMinedTableData(address, "", "")
		if err != nil {
			return err
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}

	data.Data = types.Eth1AddressPageData{
		Address:           address,
		Metadata:          metadata,
		TransactionsTable: txns,
		InternalTxnsTable: internal,
		Erc20Table:        erc20,
		Erc721Table:       erc721,
		Erc1155Table:      erc1155,
		BlocksMinedTable:  blocksMined,
		UnclesMinedTable:  unclesMined,
	}

	if utils.Config.Frontend.Debug {
		eth1AddressTemplate = template.Must(template.New("address").Funcs(utils.GetTemplateFuncs()).ParseFiles("templates/layout.html", "templates/execution/address.html"))
	}

	err = eth1AddressTemplate.ExecuteTemplate(w, "layout", data)
	if err != nil {
		logger.Errorf("error executing template for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)
	addressBytes := common.FromHex(address)

	pageToken := q.Get("pageToken")

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressTransactionsTableData(addressBytes, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressBlocksMined(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken := q.Get("pageToken")

	search := ""
	data, err := db.BigtableClient.GetAddressBlocksMinedTableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressUnclesMined(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken := q.Get("pageToken")

	search := ""
	data, err := db.BigtableClient.GetAddressUnclesMinedTableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressInternalTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)
	addressBytes := common.FromHex(address)

	pageToken := q.Get("pageToken")

	search := ""

	data, err := db.BigtableClient.GetAddressInternalTableData(addressBytes, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressErc20Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	addressBytes := common.FromHex(address)
	pageToken := q.Get("pageToken")

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressErc20TableData(addressBytes, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 internal transactions table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressErc721Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)

	pageToken := q.Get("pageToken")
	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressErc721TableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 block table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1AddressErc1155Transactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	q := r.URL.Query()
	vars := mux.Vars(r)
	address := strings.Replace(vars["address"], "0x", "", -1)
	address = strings.ToLower(address)
	pageToken := q.Get("pageToken")

	search := ""
	// logger.Infof("GETTING TRANSACTION table data for address: %v search: %v draw: %v start: %v length: %v", address, search, draw, start, length)
	data, err := db.BigtableClient.GetAddressErc1155TableData(address, search, pageToken)
	if err != nil {
		logger.WithError(err).Errorf("error getting eth1 internal transactions table data")
	}

	// logger.Infof("GOT TX: %+v", data)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		logger.Errorf("error enconding json response for %v route: %v", r.URL.String(), err)
		http.Error(w, "Internal server error", http.StatusServiceUnavailable)
		return
	}
}

func Eth1Transaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	transactionHash := vars["hash"]

	http.Redirect(w, r, fmt.Sprintf("https://etherscan.io/tx/%s", transactionHash), http.StatusSeeOther)
}

func Eth1Block(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	block := vars["block"]

	http.Redirect(w, r, fmt.Sprintf("https://etherscan.io/block/%s", block), http.StatusSeeOther)
}
