package piot_test

import (
    "testing"
    "github.com/mnezerka/go-piot"
    "github.com/mnezerka/go-piot/test"
)

func TestPrimitiveToString(t *testing.T) {

    // integer
    str, err := piot.PrimitiveToString(10)
    test.Ok(t, err)
    test.Equals(t, "10", str)

    // float
    str, err = piot.PrimitiveToString(10.23)
    test.Ok(t, err)
    test.Equals(t, "10.23", str)

    // string
    str, err = piot.PrimitiveToString("hello")
    test.Ok(t, err)
    test.Equals(t, "hello", str)

    // boolean
    str, err = piot.PrimitiveToString(true)
    test.Ok(t, err)
    test.Equals(t, "true", str)
}
