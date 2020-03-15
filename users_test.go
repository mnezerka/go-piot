package piot_test

import (
    "testing"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/test"
)

func TestFindUserByNotExistingEmail(t *testing.T) {
    db := test.GetDb(t)
    log := test.GetLogger(t)
    users := piot.NewUsers(log, db)

    test.CleanDb(t, db)
    _, err := users.FindByEmail("xx")
    test.Assert(t, err != nil, "User shall not be found")
}

func TestFindUserByExistingEmail(t *testing.T) {
    db := test.GetDb(t)
    log := test.GetLogger(t)
    users := piot.NewUsers(log, db)

    test.CleanDb(t, db)
    userId := test.CreateUser(t, db, "test1@com", "pass")
    orgId := test.CreateOrg(t, db, "testorg")
    test.AddOrgUser(t, db, orgId, userId)

    user, err := users.FindByEmail("test1@com")
    test.Ok(t, err)
    test.Equals(t, "test1@com", user.Email)
    test.Equals(t, 1, len(user.Orgs))
    test.Equals(t, "testorg", user.Orgs[0].Name)
}
