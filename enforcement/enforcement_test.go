// Copyright 2021 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enforcement

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

var (
	testData = "testdata/"
)

func TestDedupe(t *testing.T) {
	tests := []struct {
		desc string
		in   Enforcements
		want Enforcements
	}{
		{
			"no dups",
			Enforcements{Required: []string{"4018073", "67891011"}},
			Enforcements{Required: []string{"4018073", "67891011"}},
		},
		{
			"no dups and empty",
			Enforcements{Required: []string{}},
			Enforcements{Required: []string{}},
		},
		{
			"with dup requireds",
			Enforcements{Required: []string{"4018073", "67891011", "4018073", "4018073"}},
			Enforcements{Required: []string{"4018073", "67891011"}},
		},
		{
			"with dup hidden",
			Enforcements{Hidden: []string{"4018073", "67891011", "4018073", "4018073"}},
			Enforcements{Hidden: []string{"4018073", "67891011"}},
		},
		{
			"with dup excluded drivers",
			Enforcements{ExcludedDrivers: []DriverExclude{
				{DriverClass: "Dupe"},
				{DriverClass: "Unique"},
				{UpdateID: "DupeID"},
				{DriverClass: "Dupe"},
				{UpdateID: "DupeID"},
				{DriverClass: "Dupe", UpdateID: "DupeID"},
				{DriverClass: "Dupe", UpdateID: "DupeID"},
			}},
			Enforcements{ExcludedDrivers: []DriverExclude{
				{DriverClass: "Dupe"},
				{DriverClass: "Unique"},
				{UpdateID: "DupeID"},
				{DriverClass: "Dupe", UpdateID: "DupeID"},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			tt.in.dedupe()
			if diff := cmp.Diff(tt.in, tt.want, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("enforcements(%s) returned unexpected diff (-want +got):\n%s", tt.desc, diff)
			}
		})
	}
}

func TestEnforcements(t *testing.T) {
	tests := []struct {
		in      string
		want    Enforcements
		wantErr error
	}{
		{"required.json",
			Enforcements{Required: []string{"4018073", "67891011"}},
			nil,
		},
		{"hidden.json",
			Enforcements{Hidden: []string{"4018073", "67891011"}},
			nil,
		},
		{"excluded-drivers.json",
			Enforcements{ExcludedDrivers: []DriverExclude{
				{DriverClass: "UnitTest"},
				{UpdateID: "deadbeef-dead-beef-dead-beefdeadbeef"},
				{
					DriverClass: "OtherSnacks",
					UpdateID:    "cafef00d-cafe-f00d-cafe-f00dcafef00d",
				},
			}},
			nil,
		},
		{"invalid.json",
			Enforcements{},
			errParsing,
		},
		{"missing.json",
			Enforcements{},
			errInvalidFile,
		},
		{"wrong-filetype.txt",
			Enforcements{},
			errFileType,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := enforcements(filepath.Join(testData, tt.in))
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("enforcements(%s) returned unexpected diff (-want +got):\n%s", tt.in, diff)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("enforcements(%s) returned unexpected error %v", tt.in, err)
			}
		})
	}
}
