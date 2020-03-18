package piot_test

import (
    "strings"
    "testing"
    "time"
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
func TestInfluxDbPushMeasurementForSensor(t *testing.T) {
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
    test.Equals(t, 1, len(httpClient.Calls))

    // check call parameters
    test.Equals(t, "http://uri/write?db=db", httpClient.Calls[0].Url)
    // cannot use next line - order of fields and tags isn't guarnteed in golang maps
    //test.Equals(t, "sensor,id=" + sensorId.Hex() + ",name=SensorAddr,class=temperature value=23", httpClient.Calls[0].Body)
    test.Contains(t, httpClient.Calls[0].Body, "sensor")
    test.Assert(t, strings.Contains(httpClient.Calls[0].Body, "sensor"), "Body doesn't contain sensor")
    test.Assert(t, strings.Contains(httpClient.Calls[0].Body, "id=" + sensorId.Hex()), "Body doesn't contain id")
    test.Assert(t, strings.Contains(httpClient.Calls[0].Body, "name=SensorAddr"), "Body doesn't contain device name")
    test.Assert(t, strings.Contains(httpClient.Calls[0].Body, "class=temperature"), "Body doesn't contain temperature")

    test.Equals(t, "user", *httpClient.Calls[0].Username)
    test.Equals(t, "pass", *httpClient.Calls[0].Password)
}

// Push measurement for thing
func TestInfluxDbPushMeasurementForDevice(t *testing.T) {
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

// Push location for thing
func TestInfluxDbPushLocForThing(t *testing.T) {
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

    // change type of the thing to thing
    thing.Type = model.THING_TYPE_DEVICE

    // push measurement for thing
    influxdb.PostLocation(thing, 1.2, 56.8, 3, 4444)

    // check if http client was NOT called
    test.Equals(t, 1, len(httpClient.Calls))

    // check call parameters
    test.Equals(t, "http://uri/write?db=db", httpClient.Calls[0].Url)
    // following line cannot be used due to random sorting of maps (tags, fields)
    //test.Equals(t, "location,id=" + thingId.Hex() + ",name=device01 lat=1.2,lng=56.8 44444000000000\n", httpClient.Calls[0].Body)
    test.Contains(t, httpClient.Calls[0].Body, "location")
    test.Contains(t, httpClient.Calls[0].Body, "name=device01")
    test.Contains(t, httpClient.Calls[0].Body, "lat=1.2")
    test.Contains(t, httpClient.Calls[0].Body, "lng=56.8")
    test.Contains(t, httpClient.Calls[0].Body, "sat=3")
    test.Contains(t, httpClient.Calls[0].Body, " 4444000000000")
    test.Equals(t, "user", *httpClient.Calls[0].Username)
    test.Equals(t, "pass", *httpClient.Calls[0].Password)
}

func TestInfluxDbLineProtocolEncoding(t *testing.T) {
    fields := map[string]interface{}{"memory": 1000}
    tags := map[string]string{"hostname": "hal9000"}
    date := time.Date(2018, 3, 4, 5, 6, 7, 9, time.UTC)

    rm := piot.NewRowMetric("name", tags, fields, date)
    buf, err := rm.Encode()

    test.Ok(t, err)
    test.Equals(t, "name,hostname=hal9000 memory=1000i 1520139967000000009\n", buf.String())

    // with chars that need escaping
    fields = map[string]interface{}{"m em": 1000}
    tags = map[string]string{"h ost": "h al"}

    rm = piot.NewRowMetric("H E LLO", tags, fields, date)
    buf, err = rm.Encode()

    test.Ok(t, err)
    test.Equals(t, "H\\ E\\ LLO,h\\ ost=h\\ al m\\ em=1000i 1520139967000000009\n", buf.String())
}
