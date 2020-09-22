// +build !luminous,!mimic

package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListVolumes(t *testing.T) {
	fsa := getFSAdmin(t)

	vl, err := fsa.ListVolumes()
	assert.NoError(t, err)
	assert.Len(t, vl, 1)
	assert.Equal(t, "cephfs", vl[0])
}

func TestEnumerateVolumes(t *testing.T) {
	fsa := getFSAdmin(t)

	ve, err := fsa.EnumerateVolumes()
	assert.NoError(t, err)
	if assert.Len(t, ve, 1) {
		assert.Equal(t, "cephfs", ve[0].Name)
		assert.Equal(t, int64(1), ve[0].ID)
	}
}

// note: some of these dumps are simplified for testing purposes if we add
// general dump support these samples may need to be expanded upon.
var sampleDump1 = []byte(`
{
  "epoch": 5,
  "default_fscid": 1,
  "filesystems": [
    {
      "mdsmap": {
        "epoch": 5,
        "flags": 18,
        "ever_allowed_features": 0,
        "explicitly_allowed_features": 0,
        "created": "2020-08-31T18:37:34.657633+0000",
        "modified": "2020-08-31T18:37:36.700989+0000",
        "tableserver": 0,
        "root": 0,
        "session_timeout": 60,
        "session_autoclose": 300,
        "min_compat_client": "0 (unknown)",
        "max_file_size": 1099511627776,
        "last_failure": 0,
        "last_failure_osd_epoch": 0,
        "compat": {
          "compat": {},
          "ro_compat": {},
          "incompat": {
            "feature_1": "base v0.20",
            "feature_2": "client writeable ranges",
            "feature_3": "default file layouts on dirs",
            "feature_4": "dir inode in separate object",
            "feature_5": "mds uses versioned encoding",
            "feature_6": "dirfrag is stored in omap",
            "feature_8": "no anchor table",
            "feature_9": "file layout v2",
            "feature_10": "snaprealm v2"
          }
        },
        "max_mds": 1,
        "in": [
          0
        ],
        "up": {
          "mds_0": 4115
        },
        "failed": [],
        "damaged": [],
        "stopped": [],
        "info": {
          "gid_4115": {
            "gid": 4115,
            "name": "Z",
            "rank": 0,
            "incarnation": 4,
            "state": "up:active",
            "state_seq": 2,
            "addr": "127.0.0.1:6809/2568111595",
            "addrs": {
              "addrvec": [
                {
                  "type": "v1",
                  "addr": "127.0.0.1:6809",
                  "nonce": 2568111595
                }
              ]
            },
            "join_fscid": -1,
            "export_targets": [],
            "features": 4540138292836696000,
            "flags": 0
          }
        },
        "data_pools": [
          1
        ],
        "metadata_pool": 2,
        "enabled": true,
        "fs_name": "cephfs",
        "balancer": "",
        "standby_count_wanted": 0
      },
      "id": 1
    }
  ]
}
`)

var sampleDump2 = []byte(`
{
  "epoch": 5,
  "default_fscid": 1,
  "filesystems": [
    {
      "mdsmap": {
        "fs_name": "wiffleball",
        "standby_count_wanted": 0
      },
      "id": 1
    },
    {
      "mdsmap": {
        "fs_name": "beanbag",
        "standby_count_wanted": 0
      },
      "id": 2
    }
  ]
}
`)

func TestParseDumpToIdents(t *testing.T) {
	R := newResponse
	fakePrefix := dumpOkPrefix + " 5"
	t.Run("error", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(nil, "", errors.New("boop")))
		assert.Error(t, err)
		assert.Equal(t, "boop", err.Error())
		assert.Nil(t, idents)
	})
	t.Run("badStatus", func(t *testing.T) {
		_, err := parseDumpToIdents(R(sampleDump1, "unexpected!", nil))
		assert.Error(t, err)
	})
	t.Run("oneVolOk", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(sampleDump1, fakePrefix, nil))
		assert.NoError(t, err)
		if assert.Len(t, idents, 1) {
			assert.Equal(t, "cephfs", idents[0].Name)
			assert.Equal(t, int64(1), idents[0].ID)
		}
	})
	t.Run("twoVolOk", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(sampleDump2, fakePrefix, nil))
		assert.NoError(t, err)
		if assert.Len(t, idents, 2) {
			assert.Equal(t, "wiffleball", idents[0].Name)
			assert.Equal(t, int64(1), idents[0].ID)
			assert.Equal(t, "beanbag", idents[1].Name)
			assert.Equal(t, int64(2), idents[1].ID)
		}
	})
	t.Run("unexpectedStatus", func(t *testing.T) {
		idents, err := parseDumpToIdents(R(sampleDump1, "slip-up", nil))
		assert.Error(t, err)
		assert.Nil(t, idents)
	})
}