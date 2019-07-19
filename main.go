package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/alertmanager/template"
)

type responseJSON struct {
	Status  int
	Message string
}

func pushBullet(alert template.Alert) {
	log.Printf("Sending message to PushBullet")

	pushBulletAPIAddr := "https://api.pushbullet.com/v2/pushes"
	if os.Getenv("PUSHBULLETAPIADDR") != "" {
		pushBulletAPIAddr = os.Getenv("PUSHBULLETAPIADDR")
	}

	pushBulletChannelTag := "santaclausgoeswild"
	if os.Getenv("PUSHBULLETCHANNELTAG") != "" {
		pushBulletAPIAddr = os.Getenv("PUSHBULLETCHANNELTAG")
	}

	pushBulletAPIToken := ""
	if os.Getenv("PUSHBULLETAPITOKEN") != "" {
		pushBulletAPIToken = os.Getenv("PUSHBULLETAPITOKEN")
	}

	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig.InsecureSkipVerify = true
	
	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
		Transport: tr,
	}
	
	requestBody, err := json.Marshal(map[string]string{
		"body":        fmt.Sprintf("Started at %s \nStatus: %s \nSeverity: %s \nLabels %v", alert.StartsAt, alert.Status, strings.ToUpper(alert.Labels["severity"]), alert.Labels),
		"title":       "[" + strings.ToUpper(alert.Labels["severity"]) + "] " + alert.Annotations["summary"],
		"type":        "note",
		"channel_tag": pushBulletChannelTag,
	})
	if err != nil {
		log.Fatalln(err)
	}

	request, err := http.NewRequest("POST", pushBulletAPIAddr, bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Access-Token", pushBulletAPIToken)
	if err != nil {
		log.Fatalln(err)
	}

	resp, err := client.Do(request)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Response: %s", body)

}

func asJSON(w http.ResponseWriter, status int, message string) {
	data := responseJSON{
		Status:  status,
		Message: message,
	}
	bytes, _ := json.Marshal(data)
	json := string(bytes[:])

	w.WriteHeader(status)
	fmt.Fprint(w, json)
}

func webhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Godoc: https://godoc.org/github.com/prometheus/alertmanager/template#Data
	data := template.Data{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		asJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Printf("Alerts: GroupLabels=%v, CommonLabels=%v", data.GroupLabels, data.CommonLabels)
	for _, alert := range data.Alerts {
		log.Printf("Alert: status=%s,Labels=%v,Annotations=%v", alert.Status, alert.Labels, alert.Annotations)

		severity := alert.Labels["severity"]
		switch strings.ToUpper(severity) {
		case "CRITICAL":
			log.Printf("Sending notification on severity: %s", severity)
			pushBullet(alert)
		case "WARNING":
			log.Printf("Skipping notification on severity: %s", severity)
			pushBullet(alert)
		default:
			log.Printf("No action on severity: %s", severity)
		}
	}

	asJSON(w, http.StatusOK, "success")
}

func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Ok!")
}

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/webhook", webhook)

	listenAddress := ":8080"
	if os.Getenv("PORT") != "" {
		listenAddress = ":" + os.Getenv("PORT")
	}

	log.Println("Starting Alertmanager Webhook Receiver v0.1")
	log.Printf("listening on: %v", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}
