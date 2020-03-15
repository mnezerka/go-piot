package piot

import (
    "github.com/op/go-logging"
    "github.com/mnezerka/go-piot/model"
)

type influxDbMockCall struct {
    Thing *model.Thing
    Value string
}

// implements IMqtt interface
type InfluxDbMock struct {
    Log *logging.Logger
    Calls []influxDbMockCall
}

func (db *InfluxDbMock) PostMeasurement(thing *model.Thing, value string) {
    db.Log.Debugf("Influxdb - post measurement, thing: %s, val: %s", thing.Name, value)
    db.Calls = append(db.Calls, influxDbMockCall{thing, value})
}

func (db *InfluxDbMock) PostSwitchState(thing *model.Thing, value string) {
    db.Log.Debugf("Influxdb - post switch state, thing: %s, val: %s", thing.Name, value)
    db.Calls = append(db.Calls, influxDbMockCall{thing, value})
}
