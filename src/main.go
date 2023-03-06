package main

import (
	"1c-bitrix-mediator/src/httpsrv"
	"1c-bitrix-mediator/src/kafka_reader"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Kuzm1ch-dev/BitrixGo"
	"github.com/gin-gonic/gin"
)

type Res struct {
}

type Manager struct {
	HttpServer1C *httpsrv.HttpServer
	KafkaReader  *kafka_reader.KafkaReader
	bitrix       *BitrixGo.Client
}

func InitManager() Manager {
	httpServer := httpsrv.InitHttpServer()
	httpServer.BuildRecordRoutes()
	KafkaReader, _ := kafka_reader.InitKafkaReader("192.168.3.77:9092", "document", 0)
	bitrixClient, _ := BitrixGo.NewClientWithWebhookAuth("https://itgrn.bitrix24.ru", 83, "a43nx0jv1srdgfkn")
	return Manager{&httpServer, &KafkaReader, bitrixClient}
}

func main() {
	manager := InitManager()

	manager.bitrix.OnTaskCreate = func(context *gin.Context) {
		log.Println("TASK CREATE")
		jsonData, _ := ioutil.ReadAll(context.Request.Body)
		str := string(jsonData)

		regexbracket := regexp.MustCompile("(%5B)|(%5D)")
		str = regexbracket.ReplaceAllString(str, "")

		regextrash := regexp.MustCompile("(&auth)|(&data)")
		str = regextrash.ReplaceAllString(str, "&")

		ss := strings.Split(str, "&")
		m := make(map[string]string)
		for _, pair := range ss {
			z := strings.Split(pair, "=")
			m[z[0]] = z[1]
		}
		log.Println(m["FIELDS_AFTERID"])
		taskid, _ := strconv.Atoi(m["FIELDS_AFTERID"])
		newTaskArray, _ := manager.bitrix.GetTask(taskid)

		var bitrixTaskMap map[string]interface{}
		err := json.Unmarshal([]byte(newTaskArray), &bitrixTaskMap)
		if err != nil {
			log.Fatal(err)
		}

		taskResult := bitrixTaskMap["result"].(map[string]interface{})
		taskMap := taskResult["task"].(map[string]interface{})

		req, err := http.NewRequest("POST", "http://srv1c.itgrn/1C_Micromanagement/hs/Bitr_Synh/TaskCreate", bytes.NewBuffer([]byte("{\"title\":"+fmt.Sprint(taskMap["title"])+"}")))
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		log.Println("response Status:", resp.Status)
		log.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("response Body:", string(body))
	}
	manager.bitrix.OnTaskEdit = func(context *gin.Context) {
		log.Println("TASK EDIT")
		log.Println(context.Request.Header)
		log.Println(context.Request.Body)
	}
	manager.bitrix.OnTaskDelete = func(context *gin.Context) {
		log.Println("TASK DELETE")
		log.Println(context.Request.Header)
		log.Println(context.Request.Body)
	}

	go manager.bitrix.Run("192.168.3.150", "8081") // HTTP Сервер для приема из Bitrix
	go manager.KafkaReader.Read()

	manager.HttpServer1C.HttpServer.POST("/TaskCreate", manager.CreateTask)

	manager.HttpServer1C.Run("192.168.3.150", "8082") // HTTP Сервер для приема из 1С
}

// Создание задачи в битрикс при получении запроса из 1С
func (m *Manager) CreateTask(context *gin.Context) {
	log.Println("Пришел запрос из 1С на создание задачи")
	log.Println(context.Request.Body)
	jsonData, err := ioutil.ReadAll(context.Request.Body)
	if err != nil {
		log.Println(err)
	}
	var task BitrixGo.Task
	json.Unmarshal(jsonData, &task)
	m.bitrix.AddTask(task)
}
