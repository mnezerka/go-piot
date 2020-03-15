package piot_test

import (
    "testing"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/model"
    "github.com/mnezerka/go-piot/test"
    "go.mongodb.org/mongo-driver/mongo"
)

func getInfluxDb(t *testing.T, db *mongo.Database, httpClient piot.IHttpClient) piot.IInfluxDb {
    log := test.GetLogger(t)
    orgs := test.GetOrgs(t, log, db)
    return piot.NewInfluxDb(log, orgs, httpClient, "http://uri", "user", "pass")
}

// Push measurement for sensor
func TestPushMeasurementForSensor(t *testing.T) {
    const DEVICE = "device01"
    const SENSOR = "SensorAddr"

    // prepare data device + sensor assigned to org
    db := test.GetDb(t)
    logger := test.GetLogger(t)
    test.CleanDb(t, db)
    test.CreateThing(t, db, DEVICE)
    sensorId := test.CreateThing(t, db, SENSOR)
    orgId := test.CreateOrg(t, db, "org1")
    test.AddOrgThing(t, db, orgId, DEVICE)
    test.AddOrgThing(t, db, orgId, SENSOR)
    httpClient := test.GetHttpClient(t, logger)
    influxdb := getInfluxDb(t, db, httpClient)
    things := test.GetThings(t, logger, db)

    // get thing instance
    thing, err := things.Get(sensorId)
    test.Ok(t, err)

    // push measurement for thing
    influxdb.PostMeasurement(thing, "23")

    // check if http client was called

    //httpClient := ctx.Value("httpclient").(*service.HttpClientMock)

    test.Equals(t, 1, len(httpClient.Calls))

    // check call parameters
    test.Equals(t, "http://uri/write?db=db", httpClient.Calls[0].Url)
    test.Equals(t, "sensor,id=" + sensorId.Hex() + ",name=SensorAddr,class=temperature value=23", httpClient.Calls[0].Body)
    test.Equals(t, "user", *httpClient.Calls[0].Username)
    test.Equals(t, "pass", *httpClient.Calls[0].Password)
}

// Push measurement for thing
func TestPushMeasurementForDevice(t *testing.T) {
    const DEVICE = "device01"

    // prepare data device + sensor assigned to org
    db := test.GetDb(t)
    logger := test.GetLogger(t)
    test.CleanDb(t, db)
    thingId := test.CreateThing(t, db, DEVICE)
    orgId := test.CreateOrg(t, db, "org1")
    test.AddOrgThing(t, db, orgId, DEVICE)
    httpClient := test.GetHttpClient(t, logger)
    influxdb := getInfluxDb(t, db, httpClient)
    things := test.GetThings(t, logger, db)

    // get thing instance
    thing, err := things.Get(thingId)
    test.Ok(t, err)

    // change type of the thing to device
    thing.Type = model.THING_TYPE_DEVICE

    // push measurement for thing
    influxdb.PostMeasurement(thing, "23")

    // check if http client was NOT called
    test.Equals(t, 0, len(httpClient.Calls))
}
