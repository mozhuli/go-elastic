package utils

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"

	//"encoding/json"
	"runtime"
	"strconv"
	"strings"

	"github.com/fatih/structs"
)

func BodyToJson(r *http.Request) map[string]interface{} {
	decoder := json.NewDecoder(r.Body)
	fmt.Println(reflect.TypeOf(r.Body).Kind())
	var dat map[string]interface{}
	err := decoder.Decode(&dat)
	if err != nil {
		panic(err)
	}
	return dat
}

func ElasticUrl() string {

	str := GetRecordSrv("elastic.service.consul")
	if str == "" {
		return "http://127.0.0.1:9200"
	}
	return str

}

func DefaultIndex() string {
	return "search"
}

func GetRecordSrv(service string) string {
	cName, addrs, err := net.LookupSRV("", "", service)
	if err != nil {
		return ""
	}
	if cName != "" {
		fmt.Println(cName)
	}
	dat1 := structs.Map(addrs[0])
	if err != nil {
		panic(err)
	}
	return "http://" + strings.Trim(dat1["Target"].(string), ".") + ":" + strconv.Itoa(int(dat1["Port"].(uint16)))

}

func GetNumCpu() int {
	num := runtime.NumCPU()
	fmt.Println(num)
	return num
}
