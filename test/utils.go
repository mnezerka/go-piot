package test

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    "runtime"
    "reflect"
    "testing"
    "time"
    "github.com/op/go-logging"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/config"
)

const LOG_FORMAT = "%{color}%{time:2006/01/02 15:04:05 -07:00 MST} [%{level:.6s}] %{shortfile} : %{color:reset}%{message}"

var db *mongo.Database
var logger *logging.Logger

// assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
    if !condition {
        _, file, line, _ := runtime.Caller(1)
        fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
        tb.FailNow()
    }
}

// ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error) {
    if err != nil {
        _, file, line, _ := runtime.Caller(1)
        fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
        tb.FailNow()
    }
}

// equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}) {
    if !reflect.DeepEqual(exp, act) {
        _, file, line, _ := runtime.Caller(1)
        fmt.Printf("\033[31m%s:%d:\n\texp: %#v\n\tgot: %#v\033[39m\n", filepath.Base(file), line, exp, act)
        tb.FailNow()
    }
}

func CleanDb(t *testing.T, db *mongo.Database) {
    db.Collection("orgs").DeleteMany(context.TODO(), bson.M{})
    db.Collection("users").DeleteMany(context.TODO(), bson.M{})
    db.Collection("orgusers").DeleteMany(context.TODO(), bson.M{})
    db.Collection("things").DeleteMany(context.TODO(), bson.M{})
    t.Log("DB is clean")
}

func CreateDevice(t *testing.T, db *mongo.Database, name string) (primitive.ObjectID) {
    res, err := db.Collection("things").InsertOne(context.TODO(), bson.M{
        "name": name,
        "piot_id": name,
        "type": "device",
        "created": int32(time.Now().Unix()),
        "enabled": true,
    })
    Ok(t, err)

    t.Logf("Created thing of type device: %v", res.InsertedID)

    return res.InsertedID.(primitive.ObjectID)
}

func CreateSwitch(t *testing.T, db *mongo.Database, name string) (primitive.ObjectID) {
    res, err := db.Collection("things").InsertOne(context.TODO(), bson.M{
        "name": name,
        "piot_id": name,
        "type": "switch",
        "created": int32(time.Now().Unix()),
        "enabled": true,
        "store_influxdb": true,
        "switch": bson.M{
            "state_topic": "state",
            "state_on": "ON",
            "state_off": "OFF",
            "command_topic": "cmnd",
            "command_on": "ON",
            "command_off": "OFF",
        },
    })
    Ok(t, err)

    t.Logf("Created thing %v", res.InsertedID)

    return res.InsertedID.(primitive.ObjectID)
}

func CreateThing(t *testing.T, db *mongo.Database, name string) (primitive.ObjectID) {
    res, err := db.Collection("things").InsertOne(context.TODO(), bson.M{
        "name": name,
        "piot_id": name,
        "type": "sensor",
        "created": int32(time.Now().Unix()),
        "enabled": true,
        "store_mysqldb": true,
        "store_influxdb": true,
        "sensor": bson.M{
            "class": "temperature",
            "measurement_topic": "value",
        },
    })
    Ok(t, err)

    t.Logf("Created thing %v", res.InsertedID)

    return res.InsertedID.(primitive.ObjectID)
}

func CreateUser(t *testing.T, db *mongo.Database, email, password string) (primitive.ObjectID) {
    hash, err := piot.GetPasswordHash(password)
    Ok(t, err)

    res, err := db.Collection("users").InsertOne(context.TODO(), bson.M{
        "email": email,
        "password": hash,
        "created": int32(time.Now().Unix()),
    })
    Ok(t, err)

    t.Logf("Created user %v", res.InsertedID)

    return res.InsertedID.(primitive.ObjectID)
}

func CreateOrg(t *testing.T, db *mongo.Database, name string) (primitive.ObjectID) {
    res, err := db.Collection("orgs").InsertOne(context.TODO(), bson.M{
        "name": name,
        "created": int32(time.Now().Unix()),
        "influxdb": "db",
        "influxdb_username": "db-username",
        "influxdb_password": "db-password",
        "mysqldb": "mysqldb",
        "mysqldb_username": "mysqldb-username",
        "mysqldb_password": "mysqldb-password",
    })
    Ok(t, err)

    t.Logf("Created org %v", res.InsertedID)

    return res.InsertedID.(primitive.ObjectID)
}

func AddOrgUser(t *testing.T, db *mongo.Database, orgId, userId primitive.ObjectID) {
    _, err := db.Collection("orgusers").InsertOne(context.TODO(), bson.M{
        "org_id": orgId,
        "user_id": userId,
        "created": int32(time.Now().Unix()),
    })
    Ok(t, err)

    t.Logf("User %v added to org %v", userId.Hex(), orgId.Hex())
}

func AddOrgThing(t *testing.T, db *mongo.Database, orgId primitive.ObjectID, thingName string) {
    _, err := db.Collection("things").UpdateOne(context.TODO(), bson.M{"name": thingName}, bson.M{"$set": bson.M{"org_id": orgId}})
    Ok(t, err)

    t.Logf("Thing %s assigned to org %s", thingName, orgId.Hex())
}

func SetSensorMeasurementTopic(t *testing.T, db *mongo.Database, thingId primitive.ObjectID, topic string) {
    _, err := db.Collection("things").UpdateOne(context.TODO(), bson.M{"_id": thingId}, bson.M{"$set": bson.M{"sensor.measurement_topic": topic}})
    Ok(t, err)
}

func SetThingTelemetryTopic(t *testing.T, db *mongo.Database, thingId primitive.ObjectID, topic string) {
    _, err := db.Collection("things").UpdateOne(context.TODO(), bson.M{"_id": thingId}, bson.M{"$set": bson.M{"telemetry_topic": topic}})
    Ok(t, err)
}

func SetThingLocationParams(
        t *testing.T,
        db *mongo.Database,
        thingId primitive.ObjectID,
        topic string,
        lat_value string,
        lng_value string,
        date_value string,
        tracking bool) {
    update := bson.M{
        "location_topic": topic,
        "location_lat_value": lat_value,
        "location_lng_value": lng_value,
        "location_date_value": date_value,
        "location_tracking": tracking,
    }

    _, err := db.Collection("things").UpdateOne(context.TODO(), bson.M{"_id": thingId}, bson.M{"$set": update})
    Ok(t, err)
}

func SetSwitchStateTopic(t *testing.T, db *mongo.Database, thingId primitive.ObjectID, topic, on, off string) {
    update := bson.M{
        "switch.state_topic": topic,
        "switch.state_on": on,
        "switch.state_off": off,
    }
    _, err := db.Collection("things").UpdateOne(context.TODO(), bson.M{"_id": thingId}, bson.M{"$set": update})
    Ok(t, err)
}

func GetConfig() *config.Parameters{
    cfg := config.NewParameters()
    cfg.LogLevel = "DEBUG"
    return cfg
}

func GetLogger(t *testing.T) *logging.Logger {

    if logger == nil {
        cfg := GetConfig()

        log, err := piot.NewLogger(LOG_FORMAT, cfg.LogLevel)
        Ok(t, err)
        logger = log
    }

    return logger
}

func GetDb(t *testing.T) *mongo.Database {

    if db == nil {

        uri := os.Getenv("MONGODB_URI")
        // try to open database
        dbClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
        Ok(t, err)

        // Check the connection
        err = dbClient.Ping(context.TODO(), nil)
        Ok(t, err)

        db = dbClient.Database("piot-test")
    }

    return db
}

func GetPiotDevices(t *testing.T, logger *logging.Logger, things *piot.Things, mqtt piot.IMqtt) *piot.PiotDevices {
    cfg := GetConfig()
    return piot.NewPiotDevices(logger, things, mqtt, cfg)
}

func GetThings(t *testing.T, logger *logging.Logger, db *mongo.Database) *piot.Things {
    return piot.NewThings(db, logger)
}

func GetMqtt(t *testing.T, logger *logging.Logger) *MqttMock {
    return &MqttMock{Log: logger}
}

func GetUsers(t *testing.T, logger *logging.Logger, db *mongo.Database) *piot.Users {
    return piot.NewUsers(logger, db)
}

func GetOrgs(t *testing.T, logger *logging.Logger, db *mongo.Database) *piot.Orgs{
    return piot.NewOrgs(logger, db)
}

func GetHttpClient(t *testing.T, logger *logging.Logger) *HttpClientMock {
    return &HttpClientMock{Log: logger}
}

func GetInfluxDb(t *testing.T, logger *logging.Logger) *InfluxDbMock {
    return &InfluxDbMock{Log: logger}
}

func GetMysqlDb(t *testing.T, logger *logging.Logger) *MysqlDbMock {
    return &MysqlDbMock{Log: logger}
}


