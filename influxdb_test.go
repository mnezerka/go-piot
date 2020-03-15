package piot_test

import (
    "testing"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/model"
    "go.mongodb.org/mongo-driver/mongo"
)

func getInfluxDb(t *testing.T, db *mongo.Database, httpClient piot.IHttpClient) piot.IInfluxDb {
    log := GetLogger(t)
    orgs := GetOrgs(t, log, db)
    return piot.NewInfluxDb(log, orgs, httpClient, "http://uri", "user", "pass")
}

// Push measurement for sensor
func TestPushMeasurementForSensor(t *testing.T) {
    const DEVICE = "device01"
    const SENSOR = "SensorAddr"

    // prepare data device + sensor assigned to org
    db := GetDb(t)
    logger := GetLogger(t)
    CleanDb(t, db)
    CreateThing(t, db, DEVICE)
    sensorId := CreateThing(t, db, SENSOR)
    orgId := CreateOrg(t, db, "org1")
    AddOrgThing(t, db, orgId, DEVICE)
    AddOrgThing(t, db, orgId, SENSOR)
    httpClient := GetHttpClient(t, logger)
    influxdb := getInfluxDb(t, db, httpClient)
    things := GetThings(t, logger, db)

    // get thing instance
    thing, err := things.Get(sensorId)
    Ok(t, err)

    // push measurement for thing
    influxdb.PostMeasurement(thing, "23")

    // check if http client was called

    //httpClient := ctx.Value("httpclient").(*service.HttpClientMock)

    Equals(t, 1, len(httpClient.Calls))

    // check call parameters
    Equals(t, "http://uri/write?db=db", httpClient.Calls[0].Url)
    Equals(t, "sensor,id=" + sensorId.Hex() + ",name=SensorAddr,class=temperature value=23", httpClient.Calls[0].Body)
    Equals(t, "user", *httpClient.Calls[0].Username)
    Equals(t, "pass", *httpClient.Calls[0].Password)
}

// Push measurement for thing
func TestPushMeasurementForDevice(t *testing.T) {
    const DEVICE = "device01"

    // prepare data device + sensor assigned to org
    db := GetDb(t)
    logger := GetLogger(t)
    CleanDb(t, db)
    thingId := CreateThing(t, db, DEVICE)
    orgId := CreateOrg(t, db, "org1")
    AddOrgThing(t, db, orgId, DEVICE)
    httpClient := GetHttpClient(t, logger)
    influxdb := getInfluxDb(t, db, httpClient)
    things := GetThings(t, logger, db)

    // get thing instance
    thing, err := things.Get(thingId)
    Ok(t, err)

    // change type of the thing to device
    thing.Type = model.THING_TYPE_DEVICE

    // push measurement for thing
    influxdb.PostMeasurement(thing, "23")

    // check if http client was NOT called
    Equals(t, 0, len(httpClient.Calls))
}
