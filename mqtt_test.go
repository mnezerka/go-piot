package piot_test

import (
    "context"
    "fmt"
    "testing"
    "github.com/op/go-logging"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "github.com/mnezerka/go-piot"
)

func getMqtt(t *testing.T, log *logging.Logger, db *mongo.Database, influxDb piot.IInfluxDb, mysqlDb piot.IMysqlDb) piot.IMqtt {
    orgs := GetOrgs(t, log, db)
    things := GetThings(t, log, db)
    return piot.NewMqtt("uri", log, things, orgs, influxDb, mysqlDb)
}

func TestMqttMsgNotSensor(t *testing.T) {

    db := GetDb(t)
    log := GetLogger(t)
    influxDb := GetInfluxDb(t, log)
    mysqlDb := GetMysqlDb(t, log)
    mqtt := getMqtt(t, log, db, influxDb, mysqlDb)

    CleanDb(t, db)

    // send message to topic that is ignored
    mqtt.ProcessMessage("xxx", "payload")

    // send message to not registered thing
    mqtt.ProcessMessage("org/hello/x", "payload")
}

func TestMqttThingTelemetry(t *testing.T) {
    const THING = "device1"
    const ORG = "org1"

    log := GetLogger(t)
    db := GetDb(t)
    influxDb := GetInfluxDb(t, log)
    mysqlDb := GetMysqlDb(t, log)
    mqtt := getMqtt(t, log, db, influxDb, mysqlDb)
    things := GetThings(t, log, db)

    CleanDb(t, db)
    thingId := CreateDevice(t, db, THING)
    SetThingTelemetryTopic(t, db, thingId, THING + "/" + "telemetry")
    orgId := CreateOrg(t, db, ORG)
    AddOrgThing(t, db, orgId, THING)

    // send telemetry message
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/telemetry", ORG, THING), "telemetry data")

    thing, err := things.Get(thingId)
    Ok(t, err)
    Equals(t, THING, thing.Name)
    Equals(t, "telemetry data", thing.Telemetry)
}

func TestMqttThingLocation(t *testing.T) {
    const THING = "device1"
    const ORG = "org1"

    log := GetLogger(t)
    db := GetDb(t)
    influxDb := GetInfluxDb(t, log)
    mysqlDb := GetMysqlDb(t, log)
    mqtt := getMqtt(t, log, db, influxDb, mysqlDb)
    things := GetThings(t, log, db)

    CleanDb(t, db)
    thingId := CreateDevice(t, db, THING)
    SetThingLocationParams(t, db, thingId , THING + "/" + "loc", "lat", "lng")
    orgId := CreateOrg(t, db, ORG)
    AddOrgThing(t, db, orgId, THING)

    // send location message
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/loc", ORG, THING), "{\"lat\": 123.234, \"lng\": 678.789}")

    thing, err := things.Get(thingId)
    Ok(t, err)
    Equals(t, THING, thing.Name)
    Assert(t, thing.Location != nil, "Thing location not initialized")
    Equals(t, 123.234, thing.Location.Latitude)
    Equals(t, 678.789, thing.Location.Longitude)
}

// incoming sensor MQTT message for registered sensor
func TestMqttMsgSensor(t *testing.T) {
    const SENSOR = "sensor1"
    const ORG = "org1"

    log := GetLogger(t)
    db := GetDb(t)
    influxDb := GetInfluxDb(t, log)
    mysqlDb := GetMysqlDb(t, log)
    mqtt := getMqtt(t, log, db, influxDb, mysqlDb)

    CleanDb(t, db)
    sensorId := CreateThing(t, db, SENSOR)
    SetSensorMeasurementTopic(t, db, sensorId, SENSOR + "/" + "value")
    orgId := CreateOrg(t, db, ORG)
    AddOrgThing(t, db, orgId, SENSOR)

    // send unit message to registered thing
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/unit", ORG, SENSOR), "C")

    // send temperature message to registered thing
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/value", ORG, SENSOR), "23")

    // check if influxdb was called
    Equals(t, 1, len(influxDb.Calls))
    Equals(t, "23", influxDb.Calls[0].Value)
    Equals(t, SENSOR, influxDb.Calls[0].Thing.Name)

    // check if mysql was called
    Equals(t, 1, len(mysqlDb.Calls))
    Equals(t, "23", mysqlDb.Calls[0].Value)
    Equals(t, SENSOR, mysqlDb.Calls[0].Thing.Name)

    // second round of calls to check proper functionality for high load
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/unit", ORG, SENSOR), "C")
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/value", ORG, SENSOR), "23")
}

// this verifies that parsing json payloads works well
func TestMqttMsgSensorWithComplexValue(t *testing.T) {
    const SENSOR = "sensor1"
    const ORG = "org1"

    log := GetLogger(t)
    db := GetDb(t)
    influxDb := GetInfluxDb(t, log)
    mysqlDb := GetMysqlDb(t, log)
    mqtt := getMqtt(t, log, db, influxDb, mysqlDb)

    CleanDb(t, db)
    sensorId := CreateThing(t, db, SENSOR)
    SetSensorMeasurementTopic(t, db, sensorId, SENSOR + "/" + "value")
    orgId := CreateOrg(t, db, ORG)
    AddOrgThing(t, db, orgId, SENSOR)

    // modify sensor thing - set value template
    _, err := db.Collection("things").UpdateOne(context.TODO(), bson.M{"_id": sensorId}, bson.M{"$set": bson.M{"sensor.measurement_value": "temp"}})
    Ok(t, err)

    // send temperature message to registered thing
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/value", ORG, SENSOR), "{\"temp\": \"23\"}")

    // check if persistent storages were called
    Equals(t, 1, len(influxDb.Calls))
    Equals(t, 1, len(mysqlDb.Calls))
    Equals(t, "23", influxDb.Calls[0].Value)
    Equals(t, SENSOR, influxDb.Calls[0].Thing.Name)
    Equals(t, "23", mysqlDb.Calls[0].Value)
    Equals(t, SENSOR, mysqlDb.Calls[0].Thing.Name)

    // more complex structure
    _, err = db.Collection("things").UpdateOne(context.TODO(), bson.M{"_id": sensorId}, bson.M{"$set": bson.M{"sensor.measurement_value": "DS18B20.Temperature"}})
    Ok(t, err)

    payload := "{\"Time\":\"2020-01-24T22:52:58\",\"DS18B20\":{\"Id\":\"0416C18091FF\",\"Temperature\":23.0}"
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/value", ORG, SENSOR), payload)

    // check if persistent storages were called
    Equals(t, 2, len(influxDb.Calls))
    Equals(t, 2, len(mysqlDb.Calls))
    Equals(t, "23", influxDb.Calls[1].Value)
    Equals(t, SENSOR, influxDb.Calls[1].Thing.Name)
    Equals(t, "23", mysqlDb.Calls[1].Value)
    Equals(t, SENSOR, mysqlDb.Calls[1].Thing.Name)
}

// test for case when more sensors share same topic
func TestMqttMsgMultipleSensors(t *testing.T) {
    const SENSOR1 = "sensor1"
    const SENSOR2 = "sensor2"
    const ORG = "org1"

    log := GetLogger(t)
    db := GetDb(t)
    influxDb := GetInfluxDb(t, log)
    mysqlDb := GetMysqlDb(t, log)
    mqtt := getMqtt(t, log, db, influxDb, mysqlDb)

    CleanDb(t, db)
    sensor1Id := CreateThing(t, db, SENSOR1)
    SetSensorMeasurementTopic(t, db, sensor1Id, "xyz/value")
    sensor2Id := CreateThing(t, db, SENSOR2)
    SetSensorMeasurementTopic(t, db, sensor2Id, "xyz/value")

    orgId := CreateOrg(t, db, ORG)
    AddOrgThing(t, db, orgId, SENSOR1)
    AddOrgThing(t, db, orgId, SENSOR2)

    // send temperature message to registered thing
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/xyz/value", ORG), "23")

    // check if persistent storages were called
    Equals(t, 2, len(influxDb.Calls))
    Equals(t, "23", influxDb.Calls[0].Value)
    Equals(t, SENSOR1, influxDb.Calls[0].Thing.Name)
    Equals(t, "23", influxDb.Calls[1].Value)
    Equals(t, SENSOR2, influxDb.Calls[1].Thing.Name)

    Equals(t, 2, len(mysqlDb.Calls))
    Equals(t, "23", mysqlDb.Calls[0].Value)
    Equals(t, SENSOR1, mysqlDb.Calls[0].Thing.Name)
    Equals(t, "23", mysqlDb.Calls[1].Value)
    Equals(t, SENSOR2, mysqlDb.Calls[1].Thing.Name)
}

func TestMqttMsgSwitch(t *testing.T) {
    const THING = "THING1"
    const ORG = "org1"

    log := GetLogger(t)
    db := GetDb(t)
    influxDb := GetInfluxDb(t, log)
    mysqlDb := GetMysqlDb(t, log)
    mqtt := getMqtt(t, log, db, influxDb, mysqlDb)

    CleanDb(t, db)
    sensorId := CreateSwitch(t, db, THING)
    SetSwitchStateTopic(t, db, sensorId, THING + "/" + "state", "ON", "OFF")
    orgId := CreateOrg(t, db, ORG)
    AddOrgThing(t, db, orgId, THING)

    // send state change to ON
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/state", ORG, THING), "ON")

    // send state change to OFF
    mqtt.ProcessMessage(fmt.Sprintf("org/%s/%s/state", ORG, THING), "OFF")

    // check if mqtt was called
    Equals(t, 2, len(influxDb.Calls))

    Equals(t, "1", influxDb.Calls[0].Value)
    Equals(t, THING, influxDb.Calls[0].Thing.Name)

    Equals(t, "0", influxDb.Calls[1].Value)
    Equals(t, THING, influxDb.Calls[1].Thing.Name)
}