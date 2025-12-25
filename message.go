package main

type Message struct {
	User     string `json:"user"`
	Text     string `json:"text"`
	Time     string `json:"time"`
	ClientID string `json:"client_id"`
}
