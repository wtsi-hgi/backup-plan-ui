package sourcesNewSchema

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
				So(r.WildcardMatch, ShouldEqual, "*")
			},
		},
		{
			name:  "set default review date",
			setup: func(r Rule) Rule { return r },
			check: func(t *testing.T, r Rule) {
				expected := time.Now().UTC().AddDate(0, DefaultReviewMonth, 0)
				So(r.ReviewAt, ShouldHappenWithin, time.Second, expected)
			},
		},
		{
			name:  "set default delete date",
			setup: func(r Rule) Rule { return r },
			check: func(t *testing.T, r Rule) {
				expected := time.Now().UTC().AddDate(0, DefaultDeleteMonth, 0)
				So(r.DeleteAt, ShouldHappenWithin, time.Second, expected)
			},
		},
		{
			name:  "set default BackupFrequency",
			setup: func(r Rule) Rule { return r },
			check: func(t *testing.T, r Rule) {
				So(r.BackupFrequency, ShouldEqual, DefaultBackupFrequency)
			},
		},
		{
			name: "not set default BackupFrequency for NoBackup",
			setup: func(r Rule) Rule {
				r.BackupType = "nobackup"

				return r
			},
			check: func(t *testing.T, r Rule) {
				So(r.BackupFrequency, ShouldEqual, 0)
			},
		},
		{
			name: "not set default BackupFrequency for manual backup",
			setup: func(r Rule) Rule {
				r.BackupType = "manual backup"

				return r
			},
			check: func(t *testing.T, r Rule) {
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
