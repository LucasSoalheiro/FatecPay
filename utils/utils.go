package utils

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"

)

type (
	ProxyHandler      func([]byte)
	OrderProxyHandler func([]byte)
	CloseHandler      func()
)

type StackTraceError struct {
	Err   error
	Trace string
}

type Pair struct {
	Key   string
	Value float64
}

type LogMessage struct {
	Timestamp time.Time
	Message   string
}

type PrintMessage struct {
	Timestamp time.Time
	Color     string
	Text      string
	Variables []interface{}
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	ConnectionCheckAddress = ""
	CheckerConn            *websocket.Conn
)

var (
	file, err    = os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	logChannel   = make(chan LogMessage, 10000)
	printChannel = make(chan PrintMessage, 10000)
)

var (
	RateLimitTrigger          bool  = true
	InitializingBotTrigger    bool  = true
	CloseOnlyTrigger          bool  = false
	SafetyTrigger             bool  = false
	MakerSafetyTrigger        bool  = false
	TempSafetyTrigger         bool  = false
	TempTriggerActive         bool  = false
	SocketDisconnectedTrigger bool  = false
	SaveToLog                       = false
	LastOpr                   int64 = TsNow
	RateLimitTimeout          int64 = 2000
)

var (
	Reset  = "\033[39m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\x1b[38;5;226m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\u001b[90m"
	White  = "\033[97m"
	Lime   = "\x1b[38;5;46m"
	Aqua   = "\x1b[38;5;159m"
	Brown  = "\x1b[38;5;130m"
	Orange = "\x1b[38;5;214m"
	Pink   = "\x1b[38;5;219m"
)

var (
	PublicHash            = Hash("public")
	HrpHash               = Hash("hrp")
	BinanceHash           = Hash("binance")
	HuobiHash             = Hash("huobi")
	JbdraxHash            = Hash("jbdrax")
	CrossxHash            = Hash("crossx")
	CrossxmakerLondonHash = Hash("crossxmakerlondon")
	CrossxmakerTokyoHash  = Hash("crossxmakertokyo")
	FineryHash            = Hash("finery")
	CrossxTokyoHash       = Hash("crossxtokyo")
	Route28Hash           = Hash("route28")
	B2C2Hash              = Hash("b2c2")
	EnigmaHash            = Hash("enigma")
	BybitHash             = Hash("bybit")
	EnclaveHash           = Hash("enclave")
	CypatorHash           = Hash("cypator")
	CypatorLondonHash     = Hash("cypatorlondon")
	Otc677Hash            = Hash("otc677")
	SpotHash              = Hash("spot")
	UsdmHash              = Hash("usdm")
	CoinmHash             = Hash("coinm")
	CrossHash             = Hash("cross")
	LondonHash            = Hash("london")
	TokyoHash             = Hash("tokyo")
	UsaHash               = Hash("usa")

	SellHash = Hash("sell")
	BuyHash  = Hash("buy")
	BTCHash  = Hash("BTC")
	ETHHash  = Hash("ETH")
	SOLHash  = Hash("SOL")
	XRPHash  = Hash("XRP")
	AVAXHash = Hash("AVAX")
)

var (
	TelegramExposureChatId = "-4093741784"
	TelegramBotToken       = "key"
	TelegramBotUrl         = "https://api.telegram.org/bot" + TelegramBotToken
)

var (
	TsNow    = time.Now().UnixNano() / int64(time.Millisecond)
	TsNowStr = fmt.Sprintf("%d", TsNow)
)

var (
	UsdtData  = make(map[string]bool)
	bestBuy   = -1000.0
	bestSell  = -1000.0
	UsdtDepeg = false
)

var customPrintLock sync.Mutex

func SendTelegramMessage(text string) {
	text = url.QueryEscape(text)
	telegramUrl := TelegramBotUrl + "/sendMessage?chat_id=" + TelegramExposureChatId + "&text=" + text
	MakeGetRequest(telegramUrl)
}

func Spread(sell float64, buy float64) float64 {
	return 10000 * (sell - buy) / sell
}

func ReverseSpread(sell float64, buy float64) float64 {
	return 10000 * (buy - sell) / buy
}

func CheckPnl(buyOrderPrice float64, sellOrderPrice float64) float64 {
	return 10000 * (sellOrderPrice - buyOrderPrice) / sellOrderPrice
}

func PrintSpreads(buy1 float64, sell1 float64, buy2 float64, sell2 float64) {
	buySpread := Spread(sell1, buy1)
	sellSpread := Spread(sell2, buy2)
	if buySpread > bestBuy && buySpread < 1000 {
		bestBuy = buySpread
	}
	if sellSpread > bestSell && sellSpread < 1000 {
		bestSell = sellSpread
	}
	fmt.Printf("buySpread: %f, sellSpread: %f, bestBuy: %f, bestSell: %f \n", buySpread, sellSpread, bestBuy, bestSell)

}

func QuadraticScale(scale float64, percentage float64) float64 {
	return math.Pow(percentage, 2) * scale
}

func Round(num float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	return math.Round(num*shift) / shift
}

func RoundUp(num float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	return math.Ceil(num*shift) / shift
}

func RoundDown(num float64, decimalPlaces int) float64 {
	shift := math.Pow(10, float64(decimalPlaces))
	return math.Floor(num*shift) / shift
}

func FromDecimalPlacesToInt(num float64) int {
	str := strconv.FormatFloat(num, 'f', -1, 64)
	parts := strings.Split(str, ".")

	if len(parts) == 2 {
		return len(parts[1])
	}

	return 0
}

func FromIntToDecimalPlaces(num int) float64 {
	divisor := math.Pow10(num)
	result := 1.0 / divisor
	return result
}

func Hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func customPrintf(color string, text string, a ...interface{}) {
	customPrintLock.Lock()
	defer customPrintLock.Unlock()
	fmt.Print(color)
	if len(a) > 0 {
		fmt.Printf(text, a...)
	} else {
		fmt.Print(text)
	}
	fmt.Print(Reset)
}

func Init(safeStart bool, saveToLog bool) {
	SaveToLog = saveToLog
	SafetyTrigger = safeStart
	StartTsNow()
	go logger()
	go printer()
}

func StartTsNow() {
	go func() {
		for {
			RefreshTsNow()
			TimeAfter(5)
		}
	}()
}

func RefreshTsNow() {
	TsNow = time.Now().UnixNano() / int64(time.Millisecond)
	TsNowStr = fmt.Sprintf("%d", TsNow)
}

func TimeAfter(ms int64) {
	<-time.After(time.Duration(ms) * time.Millisecond)
}

func GetLastOprAge() int64 {
	return TsNow - LastOpr
}

func CastToMapSlice(data []interface{}) ([]map[string]interface{}, error) {
	mapSlice := make([]map[string]interface{}, len(data))

	for i, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			mapSlice[i] = m
		} else {
			return nil, fmt.Errorf("failed to cast element at index %d to map[string]interface{}", i)
		}
	}

	return mapSlice, nil
}

func MainConsoleLoop(closeFunc CloseHandler) {
	reader := bufio.NewReader(os.Stdin)
	for {
		//fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimRight(text, "\n")
		fmt.Println("You entered:", text)
		if text == "exit" {

			if closeFunc != nil {
				closeFunc()
			}
			AsyncPrint("yellow", "Stopping bot")
			return
		}
		if text == "stop" {
			AsyncPrint("red", "Activating safety trigger")
			SafetyTrigger = true
		}

		if text == "start" {
			AsyncPrint("green", "Deactivating safety trigger")
			SafetyTrigger = false
		}
	}
}

func ObjToJsonString(obj interface{}) string {
	jsonData, err := json.Marshal(obj)

	if err != nil {
		fmt.Println("Error converting obj to JSON:", err)
	}
	return string(jsonData)
}

func MapToJSONString(data map[string]string) (string, error) {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

func FloatToString(value float64) string {
	buffer := make([]byte, 0, 32)
	buffer = strconv.AppendFloat(buffer, value, 'f', -1, 64)
	return string(buffer)
}

func FloatToStringExact(decimalPlaces int, value float64) string {
	return fmt.Sprintf("%.*f", decimalPlaces, value)
}

func BoolToString(value bool) string {
	return strconv.FormatBool(value)
}

func StringToBool(value string) bool {
	newBool, _ := strconv.ParseBool(value)
	return newBool
}

func StringNumberToBool(value string) bool {
	return value == "0"
}

func StringToFloat(value string) float64 {
	newStr, _ := strconv.ParseFloat(value, 64)
	return newStr
}

func StringToInt(value string) int {
	newStr, _ := strconv.Atoi(value)
	return newStr
}

func IntToString(value int) string {
	return strconv.Itoa(value)
}

func UInt32ToString(value uint32) string {
	return strconv.FormatUint(uint64(value), 10)
}

func Int64ToString(value int64) string {
	return strconv.FormatInt(value, 10)
}

func StringToInt64(value string) int64 {
	number, _ := strconv.ParseInt(value, 10, 64)
	return number
}
func CoinToContract(price float64, quantity float64, contractSize float64) int {
	mult := price * quantity
	result := mult / contractSize
	return int(result)
} //OK

func ContractToCoin(price float64, quantity float64, contractSize float64) float64 {
	mult := quantity / price
	result := mult * contractSize
	return result
} //OK

func QuoteQty(price float64, quantity float64) float64 {
	mult := quantity * price
	return mult
} //OK

func SliceContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func ToResultMap(jsonStr string) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	if len(result) == 0 && jsonStr != "" {
		result = make(map[string]interface{})
		result["data"] = ToResultInterface(jsonStr)
	}
	return result
}

func ToResultInterface(jsonStr string) interface{} {
	var result interface{}
	json.Unmarshal([]byte(jsonStr), &result)
	return result
}

func ErrorHandler() {
	if err := recover(); err != nil {
		stack := string(debug.Stack())
		AsyncPrint("red", "Panic: %v", err)
		AsyncPrint("red", ", StackTrace: %s", stack)
		Log("Panic Reason: %s\n", stack)
	}
}

func InterfaceSliceToObjectSlice[T any](dataLs []interface{}) []T {
	var objLs []T
	for _, data := range dataLs {
		obj := InterfaceToObject[T](data)
		objLs = append(objLs, obj)
	}
	return objLs
}

func InterfaceSliceToObjectMap[T any](dataLs []interface{}, keyPropertyName string) map[string]T {
	objMap := make(map[string]T)
	for _, data := range dataLs {
		obj := InterfaceToObject[T](data)
		keyValue := GetPropertyValue(obj, keyPropertyName)
		key := keyValue.Interface().(string)
		objMap[key] = obj
	}
	return objMap
}

func OrderStructListByProperty[T any](list []T, propertyName string, ascending bool) []T {
	property := reflect.ValueOf(list[0]).FieldByName(propertyName).Interface()

	sort.Slice(list, func(i, j int) bool {
		val1 := reflect.ValueOf(list[i]).FieldByName(propertyName).Interface()
		val2 := reflect.ValueOf(list[j]).FieldByName(propertyName).Interface()

		switch property.(type) {
		case int:
			if ascending {
				return val1.(int) < val2.(int)

			} else {
				return val1.(int) > val2.(int)
			}
		case float64:
			if ascending {
				return val1.(float64) < val2.(float64)

			} else {
				return val1.(float64) > val2.(float64)
			}
		case string:
			if ascending {
				return val1.(string) < val2.(string)

			} else {
				return val1.(string) > val2.(string)
			}
		// Add cases for other property types if needed

		default:
			// Handle unsupported property types
			panic("Unsupported property type")
		}
	})
	return list
}

func InterfaceToObject[T any](data interface{}) T {
	var obj T
	jsonStr, _ := json.Marshal(data)
	json.Unmarshal([]byte(jsonStr), &obj)
	return obj
}

func GetPropertyValue(obj interface{}, propertyName string) reflect.Value {
	value := reflect.ValueOf(obj)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	propertyValue := value.FieldByName(propertyName)
	if propertyValue.IsValid() {
		return propertyValue
	}

	return reflect.Value{}
}

func InsertToSliceAt[T any](slice []T, index int, element T) []T {
	slice = append(slice[:index], append([]T{element}, slice[index:]...)...)
	return slice
}

func RemoveFromSliceAt[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

func RemoveFromSlice[T any](slice []T, orderToRemove T) []T {
	for i, order := range slice {
		if reflect.DeepEqual(order, orderToRemove) {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func StructExists[T any](obj T) bool {
	return reflect.ValueOf(obj).IsZero()
}

func OrderMapByValues(dataMap map[string]float64, ascending bool) map[string]float64 {

	// Extract key-value pairs into a slice

	var pairs []Pair
	for key, value := range dataMap {
		pairs = append(pairs, Pair{Key: key, Value: value})
	}

	// Sort the slice based on values
	if ascending {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value < pairs[j].Value
		})
	} else {
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Value > pairs[j].Value
		})
	}

	// Create a new sorted map
	sortedMap := make(map[string]float64)
	for _, pair := range pairs {
		sortedMap[pair.Key] = pair.Value
	}

	return sortedMap
}

func CheckMakerOrder(clientOrderId string, makerString string) (bool, string) {
	if strings.HasPrefix(clientOrderId, makerString) {
		return true, "Maker"
	}
	return false, "Taker"
}

func NewStackTraceError(err error) *StackTraceError {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stackTrace := string(buf[:n])
	return &StackTraceError{
		Err:   err,
		Trace: stackTrace,
	}
}

func (e *StackTraceError) Error() string {
	AsyncPrint("red", "%v\nStack Trace:\n%s", e.Err, e.Trace)
	return fmt.Sprintf("%v\nStack Trace:\n%s", e.Err, e.Trace)
}

func RecoverAndTryAgain(fn func()) {
	if r := recover(); r != nil {
		// Print the name of the function
		pc, _, _, _ := runtime.Caller(1)
		fnName := runtime.FuncForPC(pc).Name()

		stack := debug.Stack()
		AsyncPrint("red", "Recovered from panic in function %s: %v", fnName, r)
		AsyncPrint("red", "panic's stack: %v ", string(stack))

		time.Sleep(3 * time.Second) // Add a delay before retrying

		// Call the function again
		fn()
	}
}

func MakeGetRequest(url string) {
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		return
	}
	defer response.Body.Close()
}

func RecoverAndStop() {
	if r := recover(); r != nil {
		// Handle the panic here
		fmt.Println("Panic occurred:", r)
		debug.PrintStack()
	}
}

func Log(text string, a ...any) {
	logChannel <- LogMessage{Timestamp: time.Now(), Message: fmt.Sprintf(text, a...)}
}

func AsyncPrint(color string, text string, a ...interface{}) {
	if a != nil { // Check if variadic arguments are not nil
		printChannel <- PrintMessage{Timestamp: time.Now(), Color: color, Text: text + "\n", Variables: a}
		// print("com")
	} else {
		printChannel <- PrintMessage{Timestamp: time.Now(), Color: color, Text: text + "\n", Variables: nil}
		// print("sem")
	}
}

func logger() {
	for msg := range logChannel {
		var str string

		str = fmt.Sprintf("%s", msg.Message)
		if msg.Message != "" {
			str = fmt.Sprintf("[%s]: %s", msg.Timestamp.Format("2006-01-02T15:04:05.00000"), msg.Message)
		} else {
			str = "\n"
		}
		fmt.Fprintf(file, str)
	}
}

func printer() {
	for print := range printChannel {

		formattedTime := time.Now().Format("2006-01-02 15:04:05.00000")
		color := strings.ToUpper(print.Color)

		var colorCode string
		switch color {
		case "RED":
			colorCode = Red
		case "GREEN":
			colorCode = Green
		case "YELLOW":
			colorCode = Yellow
		case "BLUE":
			colorCode = Blue
		case "PURPLE":
			colorCode = Purple
		case "CYAN":
			colorCode = Cyan
		case "GRAY":
			colorCode = Gray
		case "WHITE":
			colorCode = White
		case "LIME":
			colorCode = Lime
		case "AQUA":
			colorCode = Aqua
		case "BROWN":
			colorCode = Brown
		case "ORANGE":
			colorCode = Orange
		case "PINK":
			colorCode = Pink
		default:
			fmt.Printf("["+formattedTime+"] "+print.Text, print.Variables)
			return
		}
		if print.Variables != nil {
			customPrintf(colorCode, "["+formattedTime+"] "+print.Text, print.Variables...)
		} else {
			customPrintf(colorCode, "["+formattedTime+"] "+print.Text)
		}

		if SaveToLog {
			Log(print.Text, print.Variables...)
		}
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MinFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func ActivateSafetyTrigger(message string) {
	oldSafetyTrigger := SafetyTrigger
	SafetyTrigger = true
	AsyncPrint("red", "@@@@ %s", message)
	if !oldSafetyTrigger {
		SendTelegramMessage(message)
	}
}

func ActivateTempSafetyTrigger() {
	TempSafetyTrigger = true
	TimeAfter(10 * 60 * 1000)
	TempSafetyTrigger = false
}
