package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var serverUrlPrefix string
var businessTypeList []string

func main() {
	consumeBusiness()
}

func init() {
	parseCommandLineArgs()
}

// 解析命令行参数
func parseCommandLineArgs() {
	flag.StringVar(&serverUrlPrefix, "server", "", "url prefix of the net request server") //参数例子: http://172.31.42.235:8080
	var businessTypeListString string
	flag.StringVar(&businessTypeListString, "business", "", "url prefix of the net request server") // 参数使用,(逗号)分隔, 参数例子: build
	flag.Parse()
	if serverUrlPrefix == "" {
		panic("请设置服务端地址")
	}
	businessTypeList = strings.Split(businessTypeListString, ",")
	log.Println("serverUrlPrefix", serverUrlPrefix)
	log.Println("businessTypeList", businessTypeList)
}

//请求网络数据
func requestNetData(url string, reqData interface{}) (map[string]interface{}, error) {
	// 请求参数
	reqDataByte, err := json.Marshal(reqData)
	// 请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqDataByte))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	// 解析响应数据
	code := resp.StatusCode
	if code != 200 {
		return nil, errors.New("request net data occur error")
	}
	//hea := resp.Header
	resDataByte, err := ioutil.ReadAll(resp.Body)
	var resData map[string]interface{}
	err = json.Unmarshal(resDataByte, &resData)
	return resData, err
}

// 同步agent参数
func syncAgentConfig() {

}

// 消费业务
func consumeBusiness() {
	log.Println("consuming business")
	defer func() {
		time.Sleep(time.Second * 1)
		recover()
		consumeBusiness()
	}()
	if len(businessTypeList) > 0 {
		resData, err := requestNetData(serverUrlPrefix+"/agent/consume_business", map[string]interface{}{
			"business_type_list": businessTypeList,
		})
		if err != nil {
			log.Println("err", err)
			return
		}
		log.Println("resData:", resData)
	}
	time.Sleep(time.Second * 1)
	consumeBusiness()
}
