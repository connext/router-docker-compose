package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var myClient = &http.Client{Timeout: 10 * time.Second}

const minRouterScrapeInterval = 60 //in seconds
const minRPCScrapeInterval = 30    //in seconds
const minETHScrapeInterval = 300   //in seconds

const RetryFetchInterval = 5 //in seconds

const metricsPreffix = "connext_"

const ETHPriceURL = "https://api.curve.fi/api/getETHprice"

//////////////////////////////////////////////////////////
//Config structure
type ConfigSt struct {
	Network              string `json:"Network"`
	Router               string `json:"Router"`
	RouterScrapeInterval int    `json:"RouterScrapeInterval"`
	RPCScrapeInterval    int    `json:"RPCScrapeInterval"`
	ETHScrapeInterval    int    `json:"ETHScrapeInterval"`
	RouterQueryLimit     int    `json:"RouterQueryLimit"`
	NetworkQueryLimit    int    `json:"NetworkQueryLimit"`
	Host                 string `json:"Host"`
	Port                 int    `json:"Port"`
}

//////////////////////////////////////////////////////////
//Router Config structures
type RouterConfigSt struct {
	Chains map[string]ChainSettingSt `json:"chains"`
}

type ChainSettingSt struct {
	Providers []string `json:"providers"`
}

//////////////////////////////////////////////////////////
//RoutersBalanceAPI endpoint structure

type RouterBalanceSt struct {
	Domain     string  `json:"domain,omitempty"` //asset_domain == domain
	AssetID    string  `json:"local,omitempty"`
	Balance    float64 `json:"balance,omitempty"`
	FeesEarned float64 `json:"fees_earned,omitempty"`
}

//RouterTransfers endpoint structures
type TransfersSt struct {
	Status           string `json:"status,omitempty"`
	OriginChain      string `json:"origin_chain,omitempty"`
	DestinationChain string `json:"destination_chain,omitempty"`
	XCallTimestamp   int32  `json:"xcall_timestamp,omitempty"`
	ExecuteTimestamp int32  `json:"execute_timestamp,omitempty"`
}

type TransfersStatusSt struct {
	Status string `json:"status,omitempty"`
}

//////////////////////////////////////////////////////////
// ETH Price structure
type ETHPriceSt struct {
	Data ETHDataSt `json:"data"`
}

type ETHDataSt struct {
	Price float64 `json:"price"`
}

//////////////////////////////////////////////////////////
// Common vars
var (
	// Endpoint variables
	Domain    = "domain"
	AssetID   = "asset"
	Status    = "status"
	RPCDomain = "RPCdomain"
	// Config variables
	router               string
	host                 string
	endpoint             string
	network              string
	port                 int
	RouterScrapeInterval int
	RPCScrapeInterval    int
	ETHScrapeInterval    int
	RouterQueryLimit     int
	NetworkQueryLimit    int

	RouterConfig RouterConfigSt
	RPCUrl       string
	eth          *ethclient.Client

	RoutersBalanceAPI = "routers_with_balances?address=eq."
	TransfersAPI      = "transfers?select=status,origin_chain,destination_chain,xcall_timestamp,execute_timestamp&order=xcall_timestamp.desc&limit="
	//TransfersStatusAPI = "transfers?select=status&routers=cs.{"

	dimensions = map[string]map[string]int{
		//// Mainnet contracts
		"6648936": map[string]int{ //Ethereum domain
			"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": 6,  //USDC local
			"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": 18, //WETH local
		},
		"1869640809": map[string]int{ //Optimism domain
			"0x67e51f46e8e14d4e4cab9df48c59ad8f512486dd": 6,  //nextUSDC local
			"0xbad5b3c68f855eaece68203312fd88ad3d365e50": 18, //nextWETH  local
		},
		"1886350457": map[string]int{ //Polygon domain
			"0xf96c6d2537e1af1a9503852eb2a4af264272a5b6": 6,  //nextUSDC local
			"0x4b8bac8dd1caa52e32c07755c17efaded6a0bbd0": 18, //nextWETH  local
		},
		"1634886255": map[string]int{ //Arbitrum-One domain
			"0x8c556cf37faa0eedac7ae665f1bb0fbd4b2eae36": 6,  //nextUSDC local
			"0x2983bf5c334743aa6657ad70a55041d720d225db": 18, //nextWETH  local
		},
		"6450786": map[string]int{ //BSC domain
			"0x5e7d83da751f4c9694b13af351b30ac108f32c38": 6,  //nextUSDC local. Actually it's 18, but Connext uses 6 in stats
			"0xa9cb51c666d2af451d87442be50747b31bb7d805": 18, //nextWETH  local
		},
		"6778479": map[string]int{ //Gnosis domain
			"0x44cf74238d840a5febb0eaa089d05b763b73fab8": 6,  //nextUSDC local
			"0x538e2ddbfdf476d24ccb1477a518a82c9ea81326": 18, //nextWETH  local
		},
		//// Testnet contracts
		"1735353714": map[string]int{ //Goerli domain
			"0x7ea6ea49b0b0ae9c5db7907d139d9cd3439862a1": 18, //TEST local
			"0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6": 18, //WETH  local
		},
		"1735356532": map[string]int{ //Optimism Goerli domain
			"0x68db1c8d85c09d546097c65ec7dcbff4d6497cbf": 18, //TEST local
			"0x39b061b7e41de8b721f9aeceb6b3f17ecb7ba63e": 18, //nextWETH  local
		},
		"9991": map[string]int{ //Mumbai (Polygon testnet) domain
			"0xedb95d8037f769b72aaab41deec92903a98c9e16": 18, //TEST local
			"0x1e5341e4b7ed5d0680d9066aac0396f0b1bd1e69": 18, //nextWETH  local
		},
		"421613": map[string]int{ //Arbitrum Goerli domain
			"0xdc805eaaabd6f68904ca706c221c72f8a8a68f9f": 18, //TEST local
			"0x1346786e6a5e07b90184a1ba58e55444b99dc4a2": 18, //nextWETH  local
		},
	}

	domains = map[string]string{
		//// Mainnet
		"6648936":    "Ethereum",
		"1869640809": "Optimism",
		"1886350457": "Polygon",
		"1634886255": "Arbitrum-One",
		"6450786":    "BSC",
		"6778479":    "Gnosis",
		//// Testnet
		"1735353714": "Goerli",
		"1735356532": "Optimism Goerli",
		"9991":       "Mumbai",
		"421613":     "Arbitrum Goerli",
	}

	assets = map[string]string{
		//// Mainnet
		"0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": "USDC",
		"0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": "WETH",
		"0x67e51f46e8e14d4e4cab9df48c59ad8f512486dd": "nextUSDC",
		"0xbad5b3c68f855eaece68203312fd88ad3d365e50": "nextWETH",
		"0xf96c6d2537e1af1a9503852eb2a4af264272a5b6": "nextUSDC",
		"0x4b8bac8dd1caa52e32c07755c17efaded6a0bbd0": "nextWETH",
		"0x8c556cf37faa0eedac7ae665f1bb0fbd4b2eae36": "nextUSDC",
		"0x2983bf5c334743aa6657ad70a55041d720d225db": "nextWETH",
		"0x5e7d83da751f4c9694b13af351b30ac108f32c38": "nextUSDC",
		"0xa9cb51c666d2af451d87442be50747b31bb7d805": "nextWETH",
		"0x44cf74238d840a5febb0eaa089d05b763b73fab8": "nextUSDC",
		"0x538e2ddbfdf476d24ccb1477a518a82c9ea81326": "nextWETH",
		//// Testnet
		"0x7ea6ea49b0b0ae9c5db7907d139d9cd3439862a1": "TEST",
		"0xb4fbf271143f4fbf7b91a5ded31805e42b2208d6": "WETH",
		"0x68db1c8d85c09d546097c65ec7dcbff4d6497cbf": "TEST",
		"0x39b061b7e41de8b721f9aeceb6b3f17ecb7ba63e": "nextWETH",
		"0xedb95d8037f769b72aaab41deec92903a98c9e16": "TEST",
		"0x1e5341e4b7ed5d0680d9066aac0396f0b1bd1e69": "nextWETH",
		"0xdc805eaaabd6f68904ca706c221c72f8a8a68f9f": "TEST",
		"0x1346786e6a5e07b90184a1ba58e55444b99dc4a2": "nextWETH",
	}
)

//////////////////////////////////////////////////////////

func LoadConfiguration(file string) ConfigSt {
	var config ConfigSt
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

func LoadRouterConfiguration(file string) RouterConfigSt {
	var config RouterConfigSt
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

//////////////////////////////////////////////////////////

// Connect to RPC server
func ConnectToRPC(url string) error {
	var err error
	eth, err = ethclient.Dial(url)
	return err
}

// Get current block
func GetBlockNumber(url string) uint64 {
	block, err := eth.BlockNumber(context.TODO())
	if err != nil {
		fmt.Printf("Error fetching current block number for %v: %v\n", url, err)
		return 0
	}
	return block
}

//////////////////////////////////////////////////////////
func GetJson(url string, target interface{}) error {
	fetched := 0

	for fetched == 0 {
		r, err := myClient.Get(url)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		//fmt.Printf("r.Body: %v\n", &r.Body)

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(target)
		if err != nil {
			fmt.Printf("Decode error: %v\n", err)
			fmt.Printf("Retyring to fetch and decode in %v seconds for URL %v\n", RetryFetchInterval, url)
			time.Sleep(time.Duration(RetryFetchInterval) * 1000 * time.Millisecond)
		} else {
			fetched = 1
		}
	}
	return nil
}

//////////////////////////////////////////////////////////

func GetRouterStats(RouterLiquidity *prometheus.GaugeVec, FeesEarned *prometheus.GaugeVec, TransfersStatus *prometheus.GaugeVec, TransfersAllStatus *prometheus.GaugeVec) error {
	//Getting router liquidity
	var cRouterBalances []RouterBalanceSt
	RouterBalanceEndpoint := endpoint + RoutersBalanceAPI + router
	//fmt.Printf("RB_URL: %v\n", RouterBalanceEndpoint)
	GetJson(RouterBalanceEndpoint, &cRouterBalances)

	//fmt.Printf("Balances: %v\n", cRouterBalances)
	//fmt.Printf("Balances[0]: %v\n", cRouterBalances[0])
	var divider_dim int
	var divider float64
	var LAsset string
	var LDomain string

	//fmt.Println(dimensions["9991"]["0xeDb95D8037f769B72AAab41deeC92903A98C9e16"])

	for i := range cRouterBalances {
		//Get dimension and calculate divider
		divider_dim = dimensions[cRouterBalances[i].Domain][cRouterBalances[i].AssetID]
		if divider_dim == 0 {
			divider_dim = 18
		}
		divider = math.Pow10(divider_dim)
		//fmt.Printf("Domain: %v  AssetID: %v  Dimension: %v\n", cRouterBalances[i].Domain, cRouterBalances[i].AssetID, divider_dim)

		//Get asset and domain names
		LAsset = assets[cRouterBalances[i].AssetID]
		if LAsset == "" {
			LAsset = cRouterBalances[i].AssetID
		}
		LDomain = domains[cRouterBalances[i].Domain]
		if LDomain == "" {
			LDomain = cRouterBalances[i].Domain
		}

		// Add info to Prometheus
		RouterLiquidity.With(prometheus.Labels{
			Domain:  LDomain,
			AssetID: LAsset,
		}).Set(cRouterBalances[i].Balance / divider)

		FeesEarned.With(prometheus.Labels{
			Domain:  LDomain,
			AssetID: LAsset,
		}).Set(cRouterBalances[i].FeesEarned / divider)
	}

	//Getting transfers statuses
	var cTransfersStatus []TransfersStatusSt
	TransfersStatusEndpoint := endpoint + TransfersAPI + strconv.Itoa(RouterQueryLimit) + "&routers=cs.{" + router + "}"
	GetJson(TransfersStatusEndpoint, &cTransfersStatus)

	// fmt.Printf("TransfersStatusEndpoint: %v\n", TransfersStatusEndpoint)
	// fmt.Printf("cTransfersStatus: %v\n", cTransfersStatus)

	Statuses := map[string]int{"XCalled": 0, "Executed": 0, "CompletedFast": 0, "CompletedSlow": 0, "Reconciled": 0}
	for i := range cTransfersStatus {
		switch cTransfersStatus[i].Status {

		case "CompletedFast":
			Statuses["CompletedFast"]++
		case "Executed":
			Statuses["Executed"]++
		case "CompletedSlow":
			Statuses["CompletedSlow"]++
		case "XCalled":
			Statuses["XCalled"]++
		case "Reconciled":
			Statuses["Reconciled"]++
		}

	}

	for key, value := range Statuses {
		TransfersStatus.With(prometheus.Labels{
			Status: key,
		}).Set(float64(value))
	}
	//fmt.Printf("XCalled: %v\n", Statuses["XCalled"])

	//Getting all network stramsfers data
	var cTransfersAll []TransfersSt
	TransfersAllEndpoint := endpoint + TransfersAPI + strconv.Itoa(NetworkQueryLimit)
	GetJson(TransfersAllEndpoint, &cTransfersAll)

	StatusesAll := map[string]int{"XCalled": 0, "Executed": 0, "CompletedFast": 0, "CompletedSlow": 0, "Reconciled": 0}
	for i := range cTransfersAll {
		switch cTransfersAll[i].Status {

		case "CompletedFast":
			StatusesAll["CompletedFast"]++
		case "Executed":
			StatusesAll["Executed"]++
		case "CompletedSlow":
			StatusesAll["CompletedSlow"]++
		case "XCalled":
			StatusesAll["XCalled"]++
		case "Reconciled":
			StatusesAll["Reconciled"]++
		}

	}

	for key, value := range StatusesAll {
		TransfersAllStatus.With(prometheus.Labels{
			Status: key,
		}).Set(float64(value))
	}

	//fmt.Printf("cTransfers: %v\n", cTransfers)

	return nil
}

//////////////////////////////////////////////////////////

func GetRPCStats(RPCStatus *prometheus.GaugeVec) error {
	var LDomain string

	for key, value := range RouterConfig.Chains {
		//fmt.Printf("Chain %v: %v \n", key, value)

		LDomain = domains[key]
		if LDomain == "" {
			LDomain = key
		}
		//fmt.Printf("Chain %v: \n", LDomain)

		for i := range value.Providers {
			RPCUrl = value.Providers[i]

			RPCerr := ConnectToRPC(RPCUrl)
			if RPCerr != nil {
				fmt.Printf("Error connecting to RPC URL %v \n", RPCUrl)
			}
			block := GetBlockNumber(RPCUrl)

			url, err := url.Parse(RPCUrl)
			if err != nil {
				log.Fatal(err)
			}

			if block != 0 { //add stats only if block != 0. 0 means error while getting stats
				RPCStatus.With(prometheus.Labels{
					Domain:    LDomain,
					RPCDomain: url.Hostname(),
				}).Set(float64(block))
			}

			//fmt.Printf("   %v  %v\n", url.Hostname(), block)
		}
	}

	return nil
}

//////////////////////////////////////////////////////////

func GetETHStats(ETHPriceStatus *prometheus.GaugeVec) error {

	var ETHPrice ETHPriceSt
	GetJson(ETHPriceURL, &ETHPrice)

	price := ETHPrice.Data.Price

	url, err := url.Parse(ETHPriceURL)
	if err != nil {
		log.Fatal(err)
	}

	if price == 0 {
		return fmt.Errorf("ETH price is 0")
	}

	ETHPriceStatus.With(prometheus.Labels{
		RPCDomain: url.Hostname(),
	}).Set(price)

	return nil
}

//////////////////////////////////////////////////////////
func main() {
	var configPath string
	var routerConfig string
	flag.StringVar(&configPath, "c", "config.json", "Specify config path")
	flag.StringVar(&routerConfig, "r", "router_config.json", "Specify router config path")
	flag.Parse()

	Config := LoadConfiguration(configPath)

	router = Config.Router
	host = Config.Host
	network = Config.Network
	port = Config.Port
	RouterScrapeInterval = Config.RouterScrapeInterval
	RPCScrapeInterval = Config.RPCScrapeInterval
	ETHScrapeInterval = Config.ETHScrapeInterval
	RouterQueryLimit = Config.RouterQueryLimit
	NetworkQueryLimit = Config.NetworkQueryLimit

	//fmt.Println("NQL is " + strconv.Itoa(NetworkQueryLimit))

	if RouterScrapeInterval < minRouterScrapeInterval {
		fmt.Printf("WARNING: minimum Router scrape Interval is %v\nSetting scrape interval to %v \n", minRouterScrapeInterval, minRouterScrapeInterval)
		RouterScrapeInterval = minRouterScrapeInterval
	}

	if RPCScrapeInterval < minRPCScrapeInterval {
		fmt.Printf("WARNING: minimum RPC scrape Interval is %v\nSetting scrape interval to %v \n", minRPCScrapeInterval, minRPCScrapeInterval)
		RPCScrapeInterval = minRPCScrapeInterval
	}

	if ETHScrapeInterval < minETHScrapeInterval {
		fmt.Printf("WARNING: minimum EHT price scrape Interval is %v\nSetting scrape interval to %v \n", minETHScrapeInterval, minETHScrapeInterval)
		ETHScrapeInterval = minETHScrapeInterval
	}

	//converting into miliseconds
	RouterScrapeInterval = RouterScrapeInterval * 1000
	RPCScrapeInterval = RPCScrapeInterval * 1000
	ETHScrapeInterval = ETHScrapeInterval * 1000

	if network == "mainnet" {
		fmt.Println("INFO: Starting in mainnet")
		endpoint = "https://postgrest.mainnet.connext.ninja/"
	} else if network == "testnet" {
		fmt.Println("INFO: Starting in testnet")
		endpoint = "https://postgrest.testnet.connext.ninja/"
	} else {
		fmt.Println("ERROR: Cannot identify network, exiting")
		os.Exit(1)
	}

	//fmt.Println("Parsing Router Config:")
	RouterConfig = LoadRouterConfiguration(routerConfig)
	//fmt.Printf("All Chains %v: \n \n", RouterConfig)

	//Custom registry
	myRegistry := prometheus.NewRegistry()

	//Prometheuse exporter settings
	RouterLiquidity := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricsPreffix + "router_liquidity",
			Help: "Router Liqudity metrics",
		}, []string{Domain, AssetID})
	myRegistry.MustRegister(RouterLiquidity)

	FeesEarned := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricsPreffix + "router_fees_earned",
			Help: "Router Fees metrics",
		}, []string{Domain, AssetID})
	myRegistry.MustRegister(FeesEarned)

	TransfersStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricsPreffix + "transfers",
			Help: "Transfers Statuses",
		}, []string{Status})
	myRegistry.MustRegister(TransfersStatus)

	TransfersAllStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricsPreffix + "transfers_all",
			Help: "Transfers Statuses for whole Connext network",
		}, []string{Status})
	myRegistry.MustRegister(TransfersAllStatus)

	RPCStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricsPreffix + "RPC_block_number",
			Help: "Latest RPC Block number",
		}, []string{Domain, RPCDomain})
	myRegistry.MustRegister(RPCStatus)

	ETHPriceStatus := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricsPreffix + "ETH_price",
			Help: "Latest ETH Price",
		}, []string{RPCDomain})
	myRegistry.MustRegister(ETHPriceStatus)

	//Loop for Router stats
	go func() {
		for {
			//fmt.Printf("Router Stats\n")
			error := GetRouterStats(RouterLiquidity, FeesEarned, TransfersStatus, TransfersAllStatus)
			if error != nil {
				fmt.Println(error)
			}
			time.Sleep(time.Duration(RouterScrapeInterval) * time.Millisecond)
		}
	}()

	//Loop for RPC stats
	go func() {
		for {
			//fmt.Printf("RPC Stats\n")
			error := GetRPCStats(RPCStatus)
			if error != nil {
				fmt.Println(error)
			}
			time.Sleep(time.Duration(RPCScrapeInterval) * time.Millisecond)
		}
	}()

	//Loop for Ethereum price stats
	go func() {
		for {
			//fmt.Printf("ETH Stats\n")
			fetched := 0
			for fetched == 0 {
				error := GetETHStats(ETHPriceStatus)
				if error != nil {
					fmt.Printf("Error fetching ETH price: %v\n", error)
					time.Sleep(time.Duration(RetryFetchInterval) * 1000 * time.Millisecond)
				} else {
					fetched = 1
				}
			}

			time.Sleep(time.Duration(ETHScrapeInterval) * time.Millisecond)
		}
	}()

	//Prometheuse exporter
	handler := promhttp.HandlerFor(myRegistry, promhttp.HandlerOpts{})
	//http.Handle("/metrics", promhttp.Handler())
	http.Handle("/metrics", handler)

	log.Printf("Starting web server at %s\n", host+":"+strconv.Itoa(port))
	err := http.ListenAndServe(host+":"+strconv.Itoa(port), nil)
	if err != nil {
		log.Printf("http.ListenAndServer: %v\n", err)
	}
}
