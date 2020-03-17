package piot_test

import (
    "context"
    "testing"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "github.com/op/go-logging"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/test"
    "github.com/mnezerka/go-piot/model"
)

type services struct {
    log *logging.Logger
    db *mongo.Database
    orgs *piot.Orgs
    things *piot.Things
    influxDb piot.IInfluxDb
    mysqlDb piot.IMysqlDb
    mqtt *test.MqttMock
    pdevices *piot.PiotDevices
}

func getServices(t *testing.T) *services {
    services := services{}
    services.log = test.GetLogger(t)
    services.db = test.GetDb(t)
    services.orgs= test.GetOrgs(t, services.log, services.db)
    services.things = test.GetThings(t, services.log, services.db)
    services.influxDb = test.GetInfluxDb(t, services.log)
    services.mysqlDb = test.GetMysqlDb(t, services.log)
    services.mqtt = test.GetMqtt(t, services.log)
    services.pdevices = test.GetPiotDevices(t, services.log, services.things, services.mqtt)

    return &services
}

// VALID packet + NEW device -> successful registration
func TestPacketDeviceReg(t *testing.T) {
    const DEVICE = "device01"
    const SENSOR = "Sensortest.Addr"

    s := getServices(t)

    test.CleanDb(t, s.db)

    // process packet for unknown device
    var packet model.PiotDevicePacket
    packet.Device = DEVICE

    var reading model.PiotSensorReading
    var temp float32 = 4.5
    reading.Address = SENSOR
    reading.Temperature = &temp
    packet.Readings = append(packet.Readings, reading)

    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // Check if device is registered
    var thing model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": DEVICE}).Decode(&thing)
    test.Ok(t, err)
    test.Equals(t, DEVICE, thing.Name)
    test.Equals(t, model.THING_TYPE_DEVICE, thing.Type)
    test.Equals(t, "available", thing.AvailabilityTopic)

    var thing_sensor model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": "T" + SENSOR}).Decode(&thing_sensor)
    test.Ok(t, err)
    test.Equals(t, "T" + SENSOR, thing_sensor.Name)
    test.Equals(t, model.THING_TYPE_SENSOR, thing_sensor.Type)
    test.Equals(t, "temperature", thing_sensor.Sensor.Class)
    test.Equals(t, "value", thing_sensor.Sensor.MeasurementTopic)

    // check correct assignment
    test.Equals(t, thing.Id, thing_sensor.ParentId)
}

// VALID packet with more measurements per 1 sensor +
// NEW device -> successful registration of device and all sensors
func TestPacketDeviceRegMultiple(t *testing.T) {
    const DEVICE = "device01"
    const SENSOR = "Sensortest.Addr"

    s := getServices(t)

    test.CleanDb(t, s.db)

    // process packet for unknown device
    var packet model.PiotDevicePacket
    packet.Device = DEVICE

    var reading model.PiotSensorReading
    reading.Address = SENSOR
    var temp float32 = 4.5
    reading.Temperature = &temp

    var press float32 = 900
    reading.Pressure = &press

    var hum float32 = 20
    reading.Humidity= &hum

    packet.Readings = append(packet.Readings, reading)

    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // Check if device is registered
    var thing model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": DEVICE}).Decode(&thing)
    test.Ok(t, err)
    test.Equals(t, DEVICE, thing.Name)
    test.Equals(t, model.THING_TYPE_DEVICE, thing.Type)
    test.Equals(t, "available", thing.AvailabilityTopic)

    var thing_sensor model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": "T" + SENSOR}).Decode(&thing_sensor)
    test.Ok(t, err)
    test.Equals(t, "T" + SENSOR, thing_sensor.Name)
    test.Equals(t, model.THING_TYPE_SENSOR, thing_sensor.Type)
    test.Equals(t, "temperature", thing_sensor.Sensor.Class)
    test.Equals(t, "value", thing_sensor.Sensor.MeasurementTopic)

    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": "P" + SENSOR}).Decode(&thing_sensor)
    test.Ok(t, err)
    test.Equals(t, "P" + SENSOR, thing_sensor.Name)
    test.Equals(t, model.THING_TYPE_SENSOR, thing_sensor.Type)
    test.Equals(t, "pressure", thing_sensor.Sensor.Class)
    test.Equals(t, "value", thing_sensor.Sensor.MeasurementTopic)

    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": "H" + SENSOR}).Decode(&thing_sensor)
    test.Ok(t, err)
    test.Equals(t, "H" + SENSOR, thing_sensor.Name)
    test.Equals(t, model.THING_TYPE_SENSOR, thing_sensor.Type)
    test.Equals(t, "humidity", thing_sensor.Sensor.Class)
    test.Equals(t, "value", thing_sensor.Sensor.MeasurementTopic)

    // check correct assignment
    test.Equals(t, thing.Id, thing_sensor.ParentId)
}


// VALID packet + NEW device -> successful registration
// VALID packet + SENSOR reassigned -> change of parent
// This test simulates scenario where sensor is disconnected
// from one device and connected to another one
func TestPacketDeviceUpdateParent(t *testing.T) {
    const DEVICE = "device01"
    const DEVICE2 = "device02"
    const SENSOR = "Sensortest.Addr"

    s := getServices(t)

    test.CleanDb(t, s.db)

    // process packet for unknown device
    var packet model.PiotDevicePacket
    packet.Device = DEVICE

    var reading model.PiotSensorReading
    var temp float32 = 4.5
    reading.Address = SENSOR
    reading.Temperature = &temp
    packet.Readings = append(packet.Readings, reading)

    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // Check if device is registered
    var thing model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": DEVICE}).Decode(&thing)
    test.Ok(t, err)
    test.Equals(t, DEVICE, thing.Name)
    test.Equals(t, model.THING_TYPE_DEVICE, thing.Type)
    test.Equals(t, "available", thing.AvailabilityTopic)

    var thing_sensor model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": "T" + SENSOR}).Decode(&thing_sensor)
    test.Ok(t, err)
    test.Equals(t, "T" + SENSOR, thing_sensor.Name)
    test.Equals(t, model.THING_TYPE_SENSOR, thing_sensor.Type)
    test.Equals(t, "temperature", thing_sensor.Sensor.Class)
    test.Equals(t, "value", thing_sensor.Sensor.MeasurementTopic)

    // check correct assignment
    test.Equals(t, thing.Id, thing_sensor.ParentId)

    // assign sensor to new device
    packet.Device = DEVICE2
    err = s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // Check if second device is registered
    var thing2 model.Thing
    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": DEVICE2}).Decode(&thing2)
    test.Ok(t, err)
    test.Equals(t, DEVICE2, thing2.Name)

    err = s.db.Collection("things").FindOne(context.TODO(), bson.M{"name": "T" + SENSOR}).Decode(&thing_sensor)
    test.Ok(t, err)
    test.Equals(t, "T" + SENSOR, thing_sensor.Name)

    // check correct re-assignment
    test.Equals(t, thing2.Id, thing_sensor.ParentId)
}


// VALID packet + UNASSIGNED device -> no mqtt messages are published
func TestPacketDeviceDataUnassigned(t *testing.T) {

    const DEVICE = "device01"

    s := getServices(t)

    test.CleanDb(t, s.db)

    // create unassigned thing
    test.CreateThing(t, s.db, DEVICE)

    // process packet for known device
    var packet model.PiotDevicePacket
    packet.Device = DEVICE

    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // check if mqtt was NOT called
    test.Equals(t, 0, len(s.mqtt.Calls))
}

// VALID packet + ASSIGNED device -> mqtt messages are published
func TestPacketDeviceDataAssigned(t *testing.T) {

    const DEVICE = "device01"

    s := getServices(t)

    test.CleanDb(t, s.db)

    // create and assign thing to org
    test.CreateThing(t, s.db, DEVICE)
    orgId := test.CreateOrg(t, s.db, "org1")
    test.AddOrgThing(t, s.db, orgId, DEVICE)

    // process packet for assigned device + provide wifi information
    var packet model.PiotDevicePacket
    packet.Device = DEVICE
    ssid := "SSID"
    packet.WifiSSID = &ssid

    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // check if mqtt was called
    test.Equals(t, 2, len(s.mqtt.Calls))

    test.Equals(t, "available", s.mqtt.Calls[0].Topic)
    test.Equals(t, "yes", s.mqtt.Calls[0].Value)

    test.Equals(t, "net/wifi/ssid", s.mqtt.Calls[1].Topic)
    test.Equals(t, "SSID", s.mqtt.Calls[1].Value)
}

// VALID packet + UNASSIGNED device + TEMPERATURE -> no mqtt messages are published
func TestPacketDeviceReadingTempUnassigned(t *testing.T) {

    const DEVICE = "device01"

    s := getServices(t)

    test.CleanDb(t, s.db)
    test.CreateThing(t, s.db, DEVICE)

    // process packet for know device
    var temp float32 = 4.5
    var reading model.PiotSensorReading
    reading.Address = "Sensortest.Addr"
    reading.Temperature = &temp

    var packet model.PiotDevicePacket
    packet.Device = DEVICE
    packet.Readings = append(packet.Readings, reading)

    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // check if mqtt was called
    test.Equals(t, 0, len(s.mqtt.Calls))
}

// VALID packet + ASSIGNED device + TEMPERATURE -> mqtt messages are published
func TestPacketDeviceReadingTempAssigned(t *testing.T) {

    const DEVICE = "device01"
    const SENSOR = "Sensortest.Addr"

    s := getServices(t)

    test.CleanDb(t, s.db)
    test.CreateThing(t, s.db, DEVICE)
    test.CreateThing(t, s.db, "T" + SENSOR)     // SENSOR is registered for temperature
    orgId := test.CreateOrg(t, s.db, "org1")
    test.AddOrgThing(t, s.db, orgId, DEVICE)
    test.AddOrgThing(t, s.db, orgId, "T" + SENSOR) // SENSOR is registered for temperature

    // process packet for know device
    var temp float32 = 4.5
    //var pressure float32 = 20.8
    //var humidity float32 = 95.5
    var reading model.PiotSensorReading
    reading.Address = SENSOR
    reading.Temperature = &temp
    //reading.Pressure= &pressure
    //reading.Humidity = &humidity

    var packet model.PiotDevicePacket
    packet.Device = DEVICE
    packet.Readings = append(packet.Readings, reading)

    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // check if mqtt was called
    test.Equals(t, 4, len(s.mqtt.Calls))

    test.Equals(t, "available", s.mqtt.Calls[0].Topic)
    test.Equals(t, "yes", s.mqtt.Calls[0].Value)
    test.Equals(t, DEVICE, s.mqtt.Calls[0].Thing.Name)

    test.Equals(t, "available", s.mqtt.Calls[1].Topic)
    test.Equals(t, "yes", s.mqtt.Calls[1].Value)
    test.Equals(t, "TSensortest.Addr", s.mqtt.Calls[1].Thing.Name)

    test.Equals(t, "value", s.mqtt.Calls[2].Topic)
    test.Equals(t, "4.5", s.mqtt.Calls[2].Value)
    test.Equals(t, "TSensortest.Addr", s.mqtt.Calls[2].Thing.Name)

    test.Equals(t, "value/unit", s.mqtt.Calls[3].Topic)
    test.Equals(t, "C", s.mqtt.Calls[3].Value)
    test.Equals(t, "TSensortest.Addr", s.mqtt.Calls[3].Thing.Name)
}

// Test DOS (Denial Of Service) protection
func TestDOS(t *testing.T) {

    s := getServices(t)

    test.CleanDb(t, s.db)

    var packet model.PiotDevicePacket

    // send first packet
    packet.Device = "device01"
    err := s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)

    // check that sending same packet in short time frame is not possible
    err = s.pdevices.ProcessPacket(packet)
    test.Assert(t, err != nil, "DOS protection doesn't work")

    // check that sending packet for different device is possible
    packet.Device = "device02"
    err = s.pdevices.ProcessPacket(packet)
    test.Ok(t, err)
}
