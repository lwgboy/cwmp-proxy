package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestReplaceConnectionUrl(t *testing.T) {
	cases := []struct {
		in, want io.Reader
	}{
		{
			bytes.NewReader([]byte(
				`
					<ParameterValueStruct>
						<Name>InternetGatewayDevice.ManagementServer.ConnectionRequestURL</Name>
						<Value xsi:type="xsd:string">http://8.8.8.8:7547</Value>
					</ParameterValueStruct>
				`,
			)),

			bytes.NewReader([]byte(
				`
					<ParameterValueStruct>
						<Name>InternetGatewayDevice.ManagementServer.ConnectionRequestURL</Name>
						<Value xsi:type="xsd:string">http://localhost:8085/client?origin=http://8.8.8.8:7547</Value>
					</ParameterValueStruct>
				`,
			)),
		},

		{
			bytes.NewReader([]byte(`<EventStruct><EventCode>0 BOOTSTRAP</EventCode><CommandKey/></EventStruct>`)),
			bytes.NewReader([]byte(`<EventStruct><EventCode>0 BOOTSTRAP</EventCode><CommandKey/></EventStruct>`)),
		},
	}

	cwmp := newCwmpMessage("localhost:8085")

	for _, c := range cases {
		got, _ := http.NewRequest("POST", "http://github.com/", c.in)
		want, _ := http.NewRequest("POST", "http://github.com/", c.want)

		cwmp.replaceConnectionUrl(got)

		compareRequests(t, want, got)
	}
}

func TestGetConnectionUrl(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{
			`
				<ParameterValueStruct>
					<Name>InternetGatewayDevice.ManagementServer.ConnectionRequestURL</Name>
					<Value xsi:type="xsd:string">http://8.8.8.8:7547</Value>
				</ParameterValueStruct>
			`,

			`http://8.8.8.8:7547`,
		},

		{
			`
				<ParameterValueStruct>
					<Name>Device.ManagementServer.ConnectionRequestURL</Name>
					<Value xsi:type="xsd:string">http://7.7.7.7:7547</Value>
				</ParameterValueStruct>
			`,

			`http://7.7.7.7:7547`,
		},
	}

	for _, c := range cases {
		got, err := getConnectionUrl(c.in)

		if err != nil {
			t.Errorf("getConnectionUrl: %v", err)
		}

		if got != c.want {
			t.Errorf("expected %v", c.want)
			t.Errorf("     got %v", got)
		}
	}
}

func TestMissingConnectionUrl(t *testing.T) {
	in := `<EventStruct><EventCode>0 BOOTSTRAP</EventCode><CommandKey/></EventStruct>`

	_, err := getConnectionUrl(in)

	if err == nil {
		t.Fatalf("getConnectionUrl: error expected, none found")
	}
}

func compareRequests(t *testing.T, want *http.Request, got *http.Request) {
	gotBuffer, _ := ioutil.ReadAll(got.Body)
	wantBuffer, _ := ioutil.ReadAll(want.Body)

	if !bytes.Equal(gotBuffer, wantBuffer) {
		t.Errorf("expected %s", wantBuffer)
		t.Errorf("     got %s", gotBuffer)
	}

	if want.ContentLength != got.ContentLength {
		t.Errorf("got length (%d) != (%d) expected length", got.ContentLength, want.ContentLength)
	}
}
