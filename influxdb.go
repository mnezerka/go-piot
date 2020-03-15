package piot

import (
    "path"
    "fmt"
    "net/url"
    "github.com/mnezerka/go-piot/model"
    "github.com/op/go-logging"
)

type IInfluxDb interface {
    PostMeasurement(thing *model.Thing, value string)
    PostSwitchState(thing *model.Thing, value string)
}

type InfluxDb struct {
    log *logging.Logger
    orgs *Orgs
    httpClient IHttpClient
    Uri string
    Username string
    Password string
}

func NewInfluxDb(log *logging.Logger, orgs *Orgs, httpClient IHttpClient, uri, username, password string) IInfluxDb {
    db := &InfluxDb{log: log, orgs: orgs, httpClient: httpClient}
    db.Uri = uri
    db.Username = username
    db.Password = password

    return db
}

func (db *InfluxDb) PostMeasurement(thing *model.Thing, value string) {
    db.log.Debugf("Posting measurement to InfluxDB, thing: %s, val: %s", thing.Name, value)

    // get thing org -> get influxdb assigned to org
    org, err := db.orgs.Get(thing.OrgId)
    if err != nil {
        return
    }

    db.log.Debugf("Going to post to InfluxDB %s as %s", org.InfluxDb, org.InfluxDbUsername)

    if thing.Type != model.THING_TYPE_SENSOR {
        // ignore things which don't represent sensor
        return
    }

    // get thing name, use alias if set
    name := thing.Name
    if thing.Alias != "" {
        name = thing.Alias
    }

    body := fmt.Sprintf("sensor,id=%s,name=%s,class=%s value=%s", thing.Id.Hex(), name, thing.Sensor.Class, value)

    url, err := url.Parse(db.Uri)
    if err != nil {
        db.log.Errorf("Cannot decode InfluxDB url from %s (%s)", db.Uri, err.Error())
        return
    }

    url.Path = path.Join(url.Path, "write")

    params := url.Query()
    params.Add("db", org.InfluxDb)
    url.RawQuery = params.Encode()

    db.httpClient.PostString(url.String(), body, &db.Username, &db.Password)
}

func (db *InfluxDb) PostSwitchState(thing *model.Thing, value string) {
    db.log.Debugf("Posting switch state to InfluxDB, thing: %s, val: %s", thing.Name, value)

    // get thing org -> get influxdb assigned to org
    org, err := db.orgs.Get(thing.OrgId)
    if err != nil {
        return
    }

    db.log.Debugf("Going to post to InfluxDB %s as %s", org.InfluxDb, org.InfluxDbUsername)

    if thing.Type != model.THING_TYPE_SWITCH {
        // ignore things which don't represent switch
        return
    }

    // get thing name, use alias if set
    name := thing.Name
    if thing.Alias != "" {
        name = thing.Alias
    }

    body := fmt.Sprintf("switch,id=%s,name=%s value=%s", thing.Id.Hex(), name, value)

    url, err := url.Parse(db.Uri)
    if err != nil {
        db.log.Errorf("Cannot decode InfluxDB url from %s (%s)", db.Uri, err.Error())
        return
    }

    url.Path = path.Join(url.Path, "write")

    params := url.Query()
    params.Add("db", org.InfluxDb)
    url.RawQuery = params.Encode()

    db.httpClient.PostString(url.String(), body, &db.Username, &db.Password)
}
