package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

// Config 保存 accesskey 的结构体
type Config struct {
	AccessKeyID     string `json:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret"`
}

// 从 JSON 配置文件中读取 accesskey
func readJSONFile(filename string) Config {
	var config Config

	f, err := os.Open(filename)
	defer f.Close()
	if err != nil {
		log.Fatalf("Faild to open the JSON file: %s", err)
	}

	dec := json.NewDecoder(f)
	if err = dec.Decode(&config); err != nil {
		log.Fatalf("Faild to parse the JSON file: %s", err)
	}

	return config
}

// 通过阿里云的 SDK 添加一条 DNS TXT 解析记录，返回记录的 RecordId，后续删除时需要用到它
func addDomainRecord(client *alidns.Client, domainName string, value string) {
	request := alidns.CreateAddDomainRecordRequest()

	request.DomainName = domainName
	request.Type = "TXT"
	request.RR = "_acme-challenge"
	request.Value = value

	response, err := client.AddDomainRecord(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("[%s] Response from 'addDomainRecord()' is %v\n", time.Now().Format("2006-01-02 15:04:05"), response)
}

// 列出所有记录类型为 TXT，且记录名包含 '_acme-challenge' 的所有记录，返回 recordID 组成的切片，后续删除它们
func listDomainRecords(client *alidns.Client, domainName string) []string {
	request := alidns.CreateDescribeDomainRecordsRequest()

	request.DomainName = domainName
	request.TypeKeyWord = "TXT"
	request.RRKeyWord = "_acme-challenge"

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("[%s] Response from 'listDomainRecords()' is %v\n", time.Now().Format("2006-01-02 15:04:05"), response)

	var recordIds []string
	for _, r := range response.DomainRecords.Record {
		recordIds = append(recordIds, r.RecordId)
	}

	return recordIds
}

// 删除解析记录
func deleteDomainRecord(client *alidns.Client, recordID string) {
	request := alidns.CreateDeleteDomainRecordRequest()

	request.RecordId = recordID

	response, err := client.DeleteDomainRecord(request)
	if err != nil {
		fmt.Print(err.Error())
	}
	fmt.Printf("[%s] Response from 'deleteDomainRecord()' is %v\n", time.Now().Format("2006-01-02 15:04:05"), response)
}

func main() {
	// 提供 -c 选项，用户可以指定JSON配置文件。注意，cfg 是一个指针
	cfg := flag.String("c", "config.json", "Assign the JSON config file")
	// 操作类型，authenticator: 域名认证，添加 DNS TXT 记录; cleanup: 认证通过后，删除此 DNS TXT 记录
	opt := flag.String("o", "authenticator", "Operate: authenticator or cleanup")
	// 提供 -h 选项，查看命令行帮助信息
	help := flag.Bool("h", false, "show help infomation")
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// 解析JSON配置文件
	config := readJSONFile(*cfg)

	// Client for Aliyun DNS SDK
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", config.AccessKeyID, config.AccessKeySecret)
	if err != nil {
		log.Fatal(err.Error())
	}

	// 判断操作类型
	switch *opt {
	case "authenticator":
		// CERTBOT_DOMAIN 和 CERTBOT_VALIDATION 是 Certbot Hooks 传过来的环境变量
		domainName := os.Getenv("CERTBOT_DOMAIN")
		value := os.Getenv("CERTBOT_VALIDATION")

		if domainName == "" || value == "" {
			log.Fatal("Error: This plugin can only be used for 'certbot' (Let's Encrypt)")
		}

		addDomainRecord(client, domainName, value)

		// Sleep to make sure the change has time to propagate over to DNS
		time.Sleep(30 * time.Second)
	case "cleanup":
		// 先获取所有记录类型为 TXT，且记录名包含 '_acme-challenge' 的记录 ID
		recordIds := listDomainRecords(client, os.Getenv("CERTBOT_DOMAIN"))
		fmt.Printf("[%s] All record Ids that need to delete is: %v\n", time.Now().Format("2006-01-02 15:04:05"), recordIds)

		// 循环，删除它们
		for _, id := range recordIds {
			deleteDomainRecord(client, id)
		}
	}
}
