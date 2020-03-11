package piot

import (
    "testing"
    "github.com/mnezerka/go-piot/model"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGetExistingThing(t *testing.T) {
    CleanDb(t)
    things := NewThings(GetDb(t), GetLogger(t))

    id := CreateThing(t, "thing1")

    thing, err := things.Get(id)
    Ok(t, err)
    Assert(t, thing.Name == "thing1", "Wrong thing name")
}


func TestGetUnknownThing(t *testing.T) {
    CleanDb(t)
    things := NewThings(GetDb(t), GetLogger(t))

    id := primitive.NewObjectID()

    _, err := things.Get(id)
    Assert(t, err != nil, "Thing shall not be found")
}

func TestFindUnknownThing(t *testing.T) {
    CleanDb(t)
    things := NewThings(GetDb(t), GetLogger(t))
    _, err := things.Find("xx")
    Assert(t, err != nil, "Thing shall not be found")
}

func TestFindExistingThing(t *testing.T) {
    CleanDb(t)
    CreateThing(t, "thing1")
    things := NewThings(GetDb(t), GetLogger(t))
    _, err := things.Find("thing1")
    Ok(t, err)
}

func TestRegisterThing(t *testing.T) {
    CleanDb(t)
    things := NewThings(GetDb(t), GetLogger(t))
    thing, err := things.RegisterPiot("thing1", "sensor")
    Ok(t, err)
    Equals(t, "thing1", thing.PiotId)
    Assert(t, thing.Name == "thing1", "Wrong thing name")
    Assert(t, thing.Type == "sensor", "Wrong thing type")
}

func TestSetParent(t *testing.T) {
    CleanDb(t)

    const THING_NAME_PARENT = "parent"
    id_parent := CreateThing(t, THING_NAME_PARENT)

    const THING_NAME_CHILD = "child"
    id_child := CreateThing(t, THING_NAME_CHILD)

    things := NewThings(GetDb(t), GetLogger(t))

    err := things.SetParent(id_child, id_parent)
    Ok(t, err)

    thing, err := things.Get(id_child)
    Ok(t, err)
    Equals(t, THING_NAME_CHILD, thing.Name)
    Equals(t, id_parent, thing.ParentId)
    /*test.Equals(t, "available", thing.AvailabilityTopic)
    test.Equals(t, "yes", thing.AvailabilityYes)
    test.Equals(t, "no", thing.AvailabilityNo)
    */
}

func TestTouchThing(t *testing.T) {
    CleanDb(t)

    const THING_NAME = "parent"
    id := CreateThing(t, THING_NAME)

    things := NewThings(GetDb(t), GetLogger(t))

    err := things.TouchThing(id)
    Ok(t, err)

    thing, err := things.Get(id)
    Ok(t, err)
    Equals(t, THING_NAME, thing.Name)
    // TODO check date
}

func TestSetAvailabilityAttributes(t *testing.T) {
    const THING_NAME = "thing2"
    CleanDb(t)
    thingId := CreateThing(t, THING_NAME)
    things := NewThings(GetDb(t), GetLogger(t))
    err := things.SetAvailabilityTopic(thingId, "available")
    Ok(t, err)
    err = things.SetAvailabilityYesNo(thingId, "yes", "no")
    Ok(t, err)

    thing, err := things.Find(THING_NAME)
    Ok(t, err)
    Equals(t, THING_NAME, thing.Name)
    Equals(t, "available", thing.AvailabilityTopic)
    Equals(t, "yes", thing.AvailabilityYes)
    Equals(t, "no", thing.AvailabilityNo)
}

func TestSetLocation(t *testing.T) {
    const THING_NAME = "thing2"
    CleanDb(t)
    thingId := CreateThing(t, THING_NAME)
    things := NewThings(GetDb(t), GetLogger(t))

    loc := model.LocationData{23.12, 56.33333};

    err := things.SetLocation(thingId, loc)
    Ok(t, err)

    thing, err := things.Find(THING_NAME)
    Ok(t, err)
    Equals(t, THING_NAME, thing.Name)
    Equals(t, 23.12, thing.Location.Latitude)
    Equals(t, 56.33333, thing.Location.Longitude)
}

func TestSetSensorAttributes(t *testing.T) {
    const THING_NAME = "thing2"
    CleanDb(t)
    thingId := CreateThing(t, THING_NAME)
    things := NewThings(GetDb(t), GetLogger(t))
    err := things.SetSensorMeasurementTopic(thingId, "value")
    Ok(t, err)

    err = things.SetSensorClass(thingId, "temperature")
    Ok(t, err)

    thing, err := things.Find(THING_NAME)
    Ok(t, err)
    Equals(t, THING_NAME, thing.Name)
    Equals(t, "value", thing.Sensor.MeasurementTopic)
    Equals(t, "temperature", thing.Sensor.Class)
}
