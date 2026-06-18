package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Endea4/studExE4-shared/events"
)

func publishDriverStatus(driverRefID, phone, gender string, isOnline bool) {
	if rdb == nil {
		return
	}
	evt := events.Event{
		Type:   events.EventDriverStatusChanged,
		Source: "user-service",
		Time:   time.Now(),
		Data: events.DriverStatusChangedData{
			DriverRefID: driverRefID,
			Phone:       phone,
			Gender:      gender,
			IsOnline:    isOnline,
		},
	}
	payload, err := json.Marshal(evt)
	if err != nil {
		logErr("publishDriverStatus marshal failed: ref=%s err=%v", driverRefID, err)
		return
	}
	if err := rdb.Publish(context.Background(), events.RedisChannelDriverStatus, payload).Err(); err != nil {
		logErr("publishDriverStatus redis publish failed: ref=%s err=%v", driverRefID, err)
	}
}
