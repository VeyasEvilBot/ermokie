package hcmmigration

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"unicode/utf16"
)

func utf16BE(value string) []byte {
	units := utf16.Encode([]rune(value))
	out := make([]byte, 2+len(units)*2)
	out[0], out[1] = 0xFE, 0xFF
	for i, unit := range units {
		binary.BigEndian.PutUint16(out[2+i*2:], unit)
	}
	return out
}

func TestLoadProfilesReturnsMalformedXMLError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.xml")
	if err := os.WriteFile(path, utf16BE("<Profiles>\n<Profile>\n</Profiles>"), 0o600); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("LoadProfiles panicked: %v", recovered)
		}
	}()
	if _, err := LoadProfiles(path); err == nil {
		t.Fatal("LoadProfiles() accepted malformed XML")
	}
}

func TestLoadProfilesAcceptsLargeProfileLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "large.xml")
	largeName := make([]byte, 128*1024)
	for i := range largeName {
		largeName[i] = 'a'
	}
	contents := "<Profiles>\n<ProfileList><Profile><Name>" + string(largeName) + "</Name></Profile></ProfileList>\n</Profiles>"
	if err := os.WriteFile(path, utf16BE(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("LoadProfiles panicked on large input: %v", recovered)
		}
	}()
	if _, err := LoadProfiles(path); err != nil {
		t.Fatalf("LoadProfiles() large input error = %v", err)
	}
}
