package sourcesnewschema

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRule(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(rule Rule) Rule
		check func(t *testing.T, rule Rule)
	}{
		{
			name:  "set default wildcard",
			setup: func(r Rule) Rule { return r },
			check: func(t *testing.T, r Rule) {
				t.Helper()

				So(r.WildcardMatch, ShouldEqual, "*")
			},
		},
		{
			name:  "set default review date",
			setup: func(r Rule) Rule { return r },
			check: func(t *testing.T, r Rule) {
				t.Helper()

				expected := time.Now().UTC().AddDate(0, DefaultReviewMonth, 0)
				So(r.ReviewAt, ShouldHappenWithin, time.Second, expected)
			},
		},
		{
			name:  "set default delete date",
			setup: func(r Rule) Rule { return r },
			check: func(t *testing.T, r Rule) {
				t.Helper()

				expected := time.Now().UTC().AddDate(0, DefaultDeleteMonth, 0)
				So(r.DeleteAt, ShouldHappenWithin, time.Second, expected)
			},
		},
		{
			name:  "set default BackupFrequency",
			setup: func(r Rule) Rule { return r },
			check: func(t *testing.T, r Rule) {
				t.Helper()

				So(r.BackupFrequency, ShouldEqual, DefaultBackupFrequency)
			},
		},
		{
			name: "not set default BackupFrequency for NoBackup",
			setup: func(r Rule) Rule {
				r.BackupType = NoBackup

				return r
			},
			check: func(t *testing.T, r Rule) {
				t.Helper()

				So(r.BackupFrequency, ShouldEqual, 0)
			},
		},
		{
			name: "not set default BackupFrequency for manual backup",
			setup: func(r Rule) Rule {
				r.BackupType = ManualBackup

				return r
			},
			check: func(t *testing.T, r Rule) {
				t.Helper()

				So(r.BackupFrequency, ShouldEqual, 0)
			},
		},
	}

	Convey("With a test rule you can", t, func() {
		baseRule := Rule{}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				rule := tc.setup(baseRule)
				rule.SetDefaults()
				tc.check(t, rule)
			})
		}
	})
}

func TestParseInstruction(t *testing.T) {
	Convey("You can parse", t, func() {
		for k, v := range instructionLookup {
			Convey(k+" instruction", func() {
				term, err := ParseInstruction(k)
				So(err, ShouldBeNil)
				So(term, ShouldEqual, v)
			})
		}
	})

	Convey("ParseInstruction will fail on incorrect instruction", t, func() {
		_, err := ParseInstruction("invalid")
		So(err, ShouldNotBeNil)
		So(err, ShouldWrap, ErrWrongInstruction)
	})
}
