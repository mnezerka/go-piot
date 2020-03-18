package piot_test

import (
    "testing"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/test"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetExistingThing(t *testing.T) {
    db := test.GetDb(t)
    test.CleanDb(t, db)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))

    id := test.CreateThing(t, db, "thing1")

    thing, err := things.Get(id)
    test.Ok(t, err)
    test.Assert(t, thing.Name == "thing1", "Wrong thing name")
}


func TestGetUnknownThing(t *testing.T) {
    db := test.GetDb(t)
    test.CleanDb(t, db)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))

    id := primitive.NewObjectID()

    _, err := things.Get(id)
    test.Assert(t, err != nil, "Thing shall not be found")
}

func TestFindUnknownThing(t *testing.T) {
    db := test.GetDb(t)
    test.CleanDb(t, db)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))
    _, err := things.Find("xx")
    test.Assert(t, err != nil, "Thing shall not be found")
}

func TestFindExistingThing(t *testing.T) {
    db := test.GetDb(t)
    test.CleanDb(t, db)
    test.CreateThing(t, db, "thing1")
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))
    _, err := things.Find("thing1")
    test.Ok(t, err)
}

func TestRegisterThing(t *testing.T) {
    db := test.GetDb(t)
    test.CleanDb(t, db)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))
    thing, err := things.RegisterPiot("thing1", "sensor")
    test.Ok(t, err)
    test.Equals(t, "thing1", thing.PiotId)
    test.Assert(t, thing.Name == "thing1", "Wrong thing name")
    test.Assert(t, thing.Type == "sensor", "Wrong thing type")
}

func TestSetParent(t *testing.T) {
    db := test.GetDb(t)
    test.CleanDb(t, db)

    const THING_NAME_PARENT = "parent"
    id_parent := test.CreateThing(t, db, THING_NAME_PARENT)

    const THING_NAME_CHILD = "child"
    id_child := test.CreateThing(t, db, THING_NAME_CHILD)

    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))

    err := things.SetParent(id_child, id_parent)
    test.Ok(t, err)

    thing, err := things.Get(id_child)
    test.Ok(t, err)
    test.Equals(t, THING_NAME_CHILD, thing.Name)
    test.Equals(t, id_parent, thing.ParentId)
    /*test.test.Equals(t, "available", thing.AvailabilityTopic)
    test.test.Equals(t, "yes", thing.AvailabilityYes)
    test.test.Equals(t, "no", thing.AvailabilityNo)
    */
}

func TestTouchThing(t *testing.T) {
    db := test.GetDb(t)
    test.CleanDb(t, db)

    const THING_NAME = "parent"
    id := test.CreateThing(t, db, THING_NAME)

    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))

    err := things.TouchThing(id)
    test.Ok(t, err)

    thing, err := things.Get(id)
    test.Ok(t, err)
    test.Equals(t, THING_NAME, thing.Name)
    // TODO check date
}

func TestSetAvailabilityAttributes(t *testing.T) {
    const THING_NAME = "thing2"
    db := test.GetDb(t)
    test.CleanDb(t, db)
    thingId := test.CreateThing(t, db, THING_NAME)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))
    err := things.SetAvailabilityTopic(thingId, "available")
    test.Ok(t, err)
    err = things.SetAvailabilityYesNo(thingId, "yes", "no")
    test.Ok(t, err)

    thing, err := things.Find(THING_NAME)
    test.Ok(t, err)
    test.Equals(t, THING_NAME, thing.Name)
    test.Equals(t, "available", thing.AvailabilityTopic)
    test.Equals(t, "yes", thing.AvailabilityYes)
    test.Equals(t, "no", thing.AvailabilityNo)
}

func TestSetLocationAttributes(t *testing.T) {
    const THING_NAME = "thing2"
    db := test.GetDb(t)
    test.CleanDb(t, db)
    thingId := test.CreateThing(t, db, THING_NAME)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))

    err := things.SetLocationMqttTopic(thingId, "loctopic")
    test.Ok(t, err)

    err = things.SetLocationMqttValues(thingId, "latval", "lngval", "satval", "tsval")
    test.Ok(t, err)

    thing, err := things.Find(THING_NAME)
    test.Ok(t, err)
    test.Equals(t, THING_NAME, thing.Name)
    test.Equals(t, "loctopic", thing.LocationMqttTopic)
    test.Equals(t, "latval", thing.LocationMqttLatValue)
    test.Equals(t, "lngval", thing.LocationMqttLngValue)
    test.Equals(t, "satval", thing.LocationMqttSatValue)
    test.Equals(t, "tsval", thing.LocationMqttTsValue)
}


func TestSetLocation(t *testing.T) {
    const THING_NAME = "thing2"
    db := test.GetDb(t)
    test.CleanDb(t, db)
    thingId := test.CreateThing(t, db, THING_NAME)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))

    err := things.SetLocation(thingId, 23.12, 56.33333, 4, 0)
    test.Ok(t, err)

    thing, err := things.Find(THING_NAME)
    test.Ok(t, err)
    test.Equals(t, THING_NAME, thing.Name)
    test.Equals(t, 23.12, thing.LocationLatitude)
    test.Equals(t, 56.33333, thing.LocationLongitude)
    test.Equals(t, int32(4), thing.LocationSatelites)
}

func TestSetSensorAttributes(t *testing.T) {
    const THING_NAME = "thing2"
    db := test.GetDb(t)
    test.CleanDb(t, db)
    thingId := test.CreateThing(t, db, THING_NAME)
    things := piot.NewThings(test.GetDb(t), test.GetLogger(t))
    err := things.SetSensorMeasurementTopic(thingId, "value")
    test.Ok(t, err)

    err = things.SetSensorClass(thingId, "temperature")
    test.Ok(t, err)

    thing, err := things.Find(THING_NAME)
    test.Ok(t, err)
    test.Equals(t, THING_NAME, thing.Name)
    test.Equals(t, "value", thing.Sensor.MeasurementTopic)
    test.Equals(t, "temperature", thing.Sensor.Class)
}
