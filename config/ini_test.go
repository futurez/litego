package config

import (
	"testing"
)

func TestIniFile(t *testing.T) {
	iniconf, err := NewConfig(IniProtocol, "config.ini")
	if err != nil {
		t.Fatal(err)
	}

	if name := iniconf.GetString("server.name"); name != "testserver" {
		t.Errorf("server.name = %s", name)
	}

	if name := iniconf.GetString("server.namedef", "testserver"); name != "testserver" {
		t.Errorf("server.namedef = %s", name)
	}

	if platform := iniconf.GetStrings("server.platform"); len(platform) != 2 {
		t.Errorf("server.platform = %q", platform)
	}

	if platform := iniconf.GetStrings("server.platformdef", "ios,android"); len(platform) != 2 {
		t.Errorf("server.platformdef = %q", platform)
	}

	if b, _ := iniconf.GetBool("server.enablessl"); !b {
		t.Errorf("server.enablessl = %t", b)
	}

	if b, _ := iniconf.GetBool("server.enablessldef", true); !b {
		t.Errorf("server.enablessldef = %t", b)
	}

	if port, _ := iniconf.GetInt("server.port"); port != 8080 {
		t.Errorf("server.port = %d", port)
	}

	if port, _ := iniconf.GetInt("server.portdef", 80); port != 80 {
		t.Errorf("server.port = %d", port)
	}

	if pi, _ := iniconf.GetFloat("server.PI"); pi != 3.14 {
		t.Errorf("server.port = %d", pi)
	}

	if pi, _ := iniconf.GetFloat("server.PIdef", 3.141); pi != 3.141 {
		t.Errorf("server.port = %d", pi)
	}
}

var iniTestData = `[server]
name = testserver
platform = android,ios
port = 8080
enablessl = true
PI = 3.14`

func TestIniData(t *testing.T) {
	iniconf, err := NewConfigData(IniProtocol, []byte(iniTestData))
	if err != nil {
		t.Fatal(err)
	}

	if name := iniconf.GetString("server.name"); name != "testserver" {
		t.Errorf("server.name = %s", name)
	}

	if name := iniconf.GetString("server.namedef", "testserver"); name != "testserver" {
		t.Errorf("server.namedef = %s", name)
	}

	if platform := iniconf.GetStrings("server.platform"); len(platform) != 2 {
		t.Errorf("server.platform = %q", platform)
	}

	if platform := iniconf.GetStrings("server.platformdef", "ios,android"); len(platform) != 2 {
		t.Errorf("server.platformdef = %q", platform)
	}

	if b, _ := iniconf.GetBool("server.enablessl"); !b {
		t.Errorf("server.enablessl = %t", b)
	}

	if b, _ := iniconf.GetBool("server.enablessldef", true); !b {
		t.Errorf("server.enablessldef = %t", b)
	}

	if port, _ := iniconf.GetInt("server.port"); port != 8080 {
		t.Errorf("server.port = %d", port)
	}

	if port, _ := iniconf.GetInt("server.portdef", 80); port != 80 {
		t.Errorf("server.port = %d", port)
	}

	if pi, _ := iniconf.GetFloat("server.PI"); pi != 3.14 {
		t.Errorf("server.port = %d", pi)
	}

	if pi, _ := iniconf.GetFloat("server.PIdef", 3.141); pi != 3.141 {
		t.Errorf("server.port = %d", pi)
	}
}
