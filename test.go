package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	ts3 "github.com/multiplay/go-ts3"
	"github.com/pkg/errors"
)

const teamspeakCheckNickname = "SinusBot via Travis CI"

type instance struct {
	UUID string `json:"uuid"`
}

func isBotRunning() error {
	if _, err := getBotID(); err != nil {
		return errors.Wrap(err, "could not get botId")
	}
	return nil
}

func canBotConnectToTeamspeak() error {
	botID, err := getBotID()
	if err != nil {
		return errors.Wrap(err, "could not get botID")
	}
	pw, err := ioutil.ReadFile(".password")
	if err != nil {
		return errors.Wrap(err, "could not read password file")
	}
	token, err := login("admin", string(pw), *botID)
	if err != nil {
		return errors.Wrap(err, "could not get token")
	}
	bots, err := getInstances(*token)
	if err != nil {
		return errors.Wrap(err, "could not get instances")
	}
	if err := changeSettings(bots[0].UUID, *token); err != nil {
		return errors.Wrap(err, "could not change instance settings")
	}
	return nil
}

func isBotOnTeamspeak() error {
	c, err := ts3.NewClient("julia.ts3index.com:10011")
	if err != nil {
		return errors.Wrap(err, "could not create new ts3 client")
	}
	defer c.Close()
	if err := c.UsePort(1489); err != nil {
		return errors.Wrap(err, "could not use port")
	}
	clientList, err := c.Server.ClientList()
	if err != nil {
		return errors.Wrap(err, "could not get clientlist")
	}
	found := false
	for _, client := range clientList {
		if strings.Contains(client.Nickname, teamspeakCheckNickname) {
			found = true
			break
		}
	}
	if !found {
		return errors.New("no client found")
	}
	return nil
}

func getInstances(token string) ([]instance, error) {
	req, err := http.NewRequest("GET", "http://127.0.0.1:8087/api/v1/bot/instances", nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not create request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not do request")
	}
	var data []instance
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, errors.Wrap(err, "could not decode json")
	}
	return data, nil
}

func changeSettings(uuid, token string) error {
	data, err := json.Marshal(map[string]string{
		"instanceId": uuid,
		"nick":       teamspeakCheckNickname,
		"serverHost": "sinusbot.com",
	})
	req, err := http.NewRequest("POST", "http://127.0.0.1:8087/api/v1/bot/i/"+uuid+"/settings", bytes.NewBuffer(data))
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not do request")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code received by setting instance settings: %d", resp.StatusCode)
	}
	req, err = http.NewRequest("POST", "http://127.0.0.1:8087/api/v1/bot/i/"+uuid+"/spawn", nil)
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "could not do request")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code received by spawning instance settings")
	}
	return nil
}

func getBotID() (*string, error) {
	resp, err := http.Get("http://127.0.0.1:8087/api/v1/botId")
	if err != nil {
		return nil, errors.Wrap(err, "could not get")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status is not expected: %d; got: %d", http.StatusOK, resp.StatusCode)
	}
	var data struct {
		DefaultBotID string `json:"defaultBotId"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, errors.Wrap(err, "could not decode data")
	}
	return &data.DefaultBotID, nil
}

func login(username, password, botID string) (*string, error) {
	data, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
		"botId":    botID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal json")
	}
	resp, err := http.Post("http://127.0.0.1:8087/api/v1/bot/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, errors.Wrap(err, "could not post")
	}
	var res struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "could not decode json")
	}
	return &res.Token, nil
}

func main() {
	fmt.Println("Checking if the bot is running...")
	if err := isBotRunning(); err != nil {
		log.Fatalf("Bot is not running: %v", err)
	}
	fmt.Println("Checking if the bot can connect to the teamspeak...")
	if err := canBotConnectToTeamspeak(); err != nil {
		log.Fatalf("Can't connect to the teamspeak server: %v", err)
	}
	fmt.Println("Sleeping so that the bot will connect in this time to the server")
	time.Sleep(5 * time.Second)
	fmt.Println("Checking if the bot is on the teamspeak")
	if err := isBotOnTeamspeak(); err != nil {
		log.Fatalf("Failed, bot is not on the teamspeak 3 server")
	}
}
