//go:build netbsd

package aghnet

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIfaceHasStaticIP(t *testing.T) {
	const (
		ifaceName   = "em0"
		rcConf      = "etc/rc.conf"
		ifconfigCfg = "etc/ifconfig." + ifaceName
		staticIP    = "inet 127.0.0.253 netmask 0xffffffff"
	)

	testCases := []struct {
		name     string
		rootFsys fs.FS
		wantHas  assert.BoolAssertionFunc
	}{
		{
			name: "rc_conf",
			rootFsys: fstest.MapFS{rcConf: &fstest.MapFile{
				Data: []byte(`ifconfig_` + ifaceName + `="` + staticIP + `"` + nl),
			}},
			wantHas: assert.True,
		},
		{
			name: "ifconfig_file",
			rootFsys: fstest.MapFS{ifconfigCfg: &fstest.MapFile{
				Data: []byte(staticIP + nl),
			}},
			wantHas: assert.True,
		},
		{
			name: "rc_conf_empty_uses_file",
			rootFsys: fstest.MapFS{
				rcConf: &fstest.MapFile{
					Data: []byte(`ifconfig_` + ifaceName + `=""` + nl),
				},
				ifconfigCfg: &fstest.MapFile{
					Data: []byte(staticIP + nl),
				},
			},
			wantHas: assert.True,
		},
		{
			name: "rc_conf_dhcp_precedes_file",
			rootFsys: fstest.MapFS{
				rcConf: &fstest.MapFile{
					Data: []byte(`ifconfig_` + ifaceName + `="dhcp"` + nl),
				},
				ifconfigCfg: &fstest.MapFile{
					Data: []byte(staticIP + nl),
				},
			},
			wantHas: assert.False,
		},
		{
			name: "rc_conf_static_precedes_file",
			rootFsys: fstest.MapFS{
				rcConf: &fstest.MapFile{
					Data: []byte(`ifconfig_` + ifaceName + `="` + staticIP + `"` + nl),
				},
				ifconfigCfg: &fstest.MapFile{
					Data: []byte("dhcp" + nl),
				},
			},
			wantHas: assert.True,
		},
		{
			name: "rc_conf_semicolon_delimited",
			rootFsys: fstest.MapFS{rcConf: &fstest.MapFile{
				Data: []byte(`ifconfig_` + ifaceName + `="dhcp; ` + staticIP + `"` + nl),
			}},
			wantHas: assert.True,
		},
		{
			name: "no_static_configuration",
			rootFsys: fstest.MapFS{ifconfigCfg: &fstest.MapFile{
				Data: []byte("dhcp" + nl),
			}},
			wantHas: assert.False,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			substRootDirFS(t, tc.rootFsys)

			ctx := testutil.ContextWithTimeout(t, testTimeout)
			has, err := IfaceHasStaticIP(ctx, testCmdCons, ifaceName)
			require.NoError(t, err)

			tc.wantHas(t, has)
		})
	}
}
