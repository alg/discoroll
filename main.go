package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.uber.org/zap"
)

type RollbarEvent struct {
	EventName string         `json:"event_name"`
	Data      map[string]any `json:"data"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func Repackage(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Could not initialize logger: %v", err)
		w.WriteHeader(500)
		return
	}
	logger := zapLogger.Sugar()

	var evt RollbarEvent
	err = json.NewDecoder(r.Body).Decode(&evt)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	logger.Infow("Rollbar event", "params", p, "name", evt.EventName, "data", evt.Data)

	discordEvent, ok, err := rollbarToDiscord(evt)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	// ignore unknown events
	if !ok {
		w.WriteHeader(200)
		return
	}

	logger.Infow("Discord event", "event", discordEvent)
	err = deliverToDiscord(discordEvent, p.ByName("webhook_id"), p.ByName("webhook_token"))
	if err != nil {
		logger.Errorf("error delivering to Discord: %v", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
}

func deliverToDiscord(discordEvent map[string]any, webhookId, webhookToken string) error {
	data, err := json.Marshal(discordEvent)
	if err != nil {
		return fmt.Errorf("error serializing discordEvent: %w", err)
	}

	url := discordUrl(webhookId, webhookToken)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("error sending webhook: %w", err)
	}

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected reply from Discord: status=%d body=%s", resp.StatusCode, body)
	}

	return nil
}

func discordUrl(webhookID, webhookToken string) string {
	return fmt.Sprintf("https://discord.com/api/webhooks/%s/%s", webhookID, webhookToken)
}

func presentNewItem(evt RollbarEvent) map[string]any {
	item := evt.Data["item"].(map[string]any)

	itemID := int64(item["counter"].(float64))
	occurrences := int(item["total_occurrences"].(float64))

	return map[string]any{
		"embeds": []any{
			map[string]any{
				"title": trimTitle(fmt.Sprintf("#%d New Error: %s", itemID, item["title"])),
				"url":   evt.Data["url"],
				"color": 12592926,
				"fields": []DiscordField{
					{Name: "Environment", Value: item["environment"].(string), Inline: true},
					{Name: "Occurences", Value: fmt.Sprintf("%v", occurrences), Inline: true},
				},
			},
		},
	}
}

func presentItemVelocity(evt RollbarEvent) map[string]any {
	item := evt.Data["item"].(map[string]any)

	itemID := int64(item["counter"].(float64))
	occurrences := int(evt.Data["occurrences"].(float64))
	trigger := evt.Data["trigger"].(map[string]any)
	windowSize := trigger["window_size_description"]

	return map[string]any{
		"embeds": []any{
			map[string]any{
				"title": trimTitle(fmt.Sprintf("#%d %d occurrences in %v: %s", itemID, occurrences, windowSize, item["title"])),
				"url":   evt.Data["url"],
				"fields": []DiscordField{
					{Name: "Environment", Value: item["environment"].(string), Inline: true},
				},
			},
		},
	}
}

func presentExpRepeatItem(evt RollbarEvent) map[string]any {
	item := evt.Data["item"].(map[string]any)

	itemID := int64(item["counter"].(float64))
	occurrences := int(evt.Data["occurrences"].(float64))

	return map[string]any{
		"embeds": []any{
			map[string]any{
				"title": trimTitle(fmt.Sprintf("#%d %dth error: %s", itemID, occurrences, item["title"])),
				"url":   evt.Data["url"],
				"fields": []DiscordField{
					{Name: "Environment", Value: item["environment"].(string), Inline: true},
				},
			},
		},
	}
}

func presentReopenedItem(evt RollbarEvent) map[string]any {
	item := evt.Data["item"].(map[string]any)

	itemID := int64(item["counter"].(float64))
	occurrences := int(item["total_occurrences"].(float64))

	return map[string]any{
		"embeds": []any{
			map[string]any{
				"title": trimTitle(fmt.Sprintf("#%d Reopened: %s", itemID, item["title"])),
				"url":   evt.Data["url"],
				"color": 12592926,
				"fields": []DiscordField{
					{Name: "Environment", Value: item["environment"].(string), Inline: true},
					{Name: "Occurrences", Value: fmt.Sprintf("%d", occurrences), Inline: true},
				},
			},
		},
	}
}

func presentResolvedItem(evt RollbarEvent) map[string]any {
	item := evt.Data["item"].(map[string]any)

	itemID := int64(item["counter"].(float64))
	occurrences := int(item["total_occurrences"].(float64))

	return map[string]any{
		"embeds": []any{
			map[string]any{
				"title": trimTitle(fmt.Sprintf("#%d Resolved: %s", itemID, item["title"])),
				"url":   evt.Data["url"],
				"color": 2015366,
				"fields": []DiscordField{
					{Name: "Environment", Value: item["environment"].(string), Inline: true},
					{Name: "Occurrences", Value: fmt.Sprintf("%v", occurrences), Inline: true},
				},
			},
		},
	}
}

const MAX_LENGTH = 247

func trimTitle(txt string) string {
	if len(txt) < MAX_LENGTH {
		return txt
	}

	return fmt.Sprintf("%s...", txt[:MAX_LENGTH])
}

func rollbarToDiscord(evt RollbarEvent) (data map[string]any, ok bool, err error) {
	switch evt.EventName {
	case "new_item":
		return presentNewItem(evt), true, nil

	case "item_velocity":
		return presentItemVelocity(evt), true, nil

	case "exp_repeat_item":
		return presentExpRepeatItem(evt), true, nil

	case "resolved_item":
		return presentResolvedItem(evt), true, nil

	case "reopened_item":
		return presentReopenedItem(evt), true, nil

	default:
		return data, false, nil
	}
}

func main() {
	router := httprouter.New()
	router.POST("/:webhook_id/:webhook_token", Repackage)

	log.Fatal(http.ListenAndServe(":8080", router))
}
