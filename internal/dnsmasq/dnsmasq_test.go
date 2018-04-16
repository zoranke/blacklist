package dnsmasq

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigFile(t *testing.T) {
	Convey("Testing ConfigFile()", t, func() {
		var (
			b     []byte
			dir   = "../testdata/etc/dnsmasq.d/"
			err   error
			files []string
			r     io.Reader
		)

		Convey("Testing with a dnsmasq entries loaded from files", func() {
			files, err = filepath.Glob(dir + "*.conf")
			So(err, ShouldBeNil)

			for _, f := range files {
				Convey("Parsing file: "+f, func() {
					if r, err = ConfigFile(f); err != nil {
						Printf("cannot open configuration file %s!", f)
					}

					b, _ = ioutil.ReadAll(r)
					c := make(Conf)
					ip := "0.0.0.0"
					So(c.Parse(&Mapping{Contents: b}), ShouldBeNil)

					for k := range c {
						So(c.Redirect(k, ip), ShouldBeTrue)
					}
				})
			}
		})

		Convey("Testing a misdirected dnsmasq address entry...", func() {
			c := make(Conf)
			ip := "0.0.0.0"
			k := "address=/www.google.com/0.0.0.0"

			So(c.Parse(&Mapping{Contents: []byte(k)}), ShouldBeNil)
			So(c.Redirect(k, ip), ShouldBeFalse)
		})
		// Convey("Testing with a disabled configuration", func() {
		// 	act := NewConfig().ReadCfg(&CFGstatic{Cfg: tdata.DisabledCfg})
		// 	So(act, ShouldBeEmpty)
		// })

		// Convey("Testing with a single source configuration", func() {
		// 	act := NewConfig().ReadCfg(&CFGstatic{Cfg: tdata.SingleSource})
		// 	So(act, ShouldBeEmpty)
		// })

		// Convey("Testing with an active configuration", func() {
		// 	c := NewConfig()
		// 	So(c.ReadCfg(&CFGstatic{Cfg: tdata.Cfg}), ShouldBeNil)
		// 	So(c.Nodes(), ShouldResemble, []string{"blacklist", "domains", "hosts"})
		// })
	})
}

func TestFetchHost(t *testing.T) {
	tests := []struct {
		conf Conf
		exp  bool
		ip   string
		key  string
		name string
	}{
		{
			ip:   "0.0.0.0",
			key:  "badguy_s.com",
			conf: Conf{"badguys.com": Host{IP: "0.0.0.0", Server: false}},
			exp:  false,
			name: "badguys.com",
		},
		{
			ip:   "127.0.0.1",
			key:  "localhoster",
			conf: Conf{"localhost": Host{IP: "127.0.0.1", Server: false}},
			exp:  false,
			name: "localhoster",
		},
		{
			ip:   "127.0.0.1",
			key:  "localhost",
			conf: Conf{"localhost": Host{IP: "#", Server: true}},
			exp:  true,
			name: "localServer",
		},
		{
			ip:   "127.0.0.1",
			key:  "localhost",
			conf: Conf{"localhost": Host{IP: "127.0.0.1", Server: false}},
			exp:  true,
			name: "localhost",
		},
		{
			ip:   "127.0.0.1",
			exp:  false,
			name: "no name",
		},
		{
			ip:   "::1",
			key:  "localhost",
			conf: Conf{"localhost": Host{IP: "127.0.0.1", Server: false}},
			exp:  false,
			name: "localhost IPv6",
		},
	}

	Convey("Testing String()", t, func() {
		for _, tt := range tests {
			Convey("current test "+tt.name, func() {
				So(fetchHost(tt.key, tt.ip), ShouldEqual, tt.exp)
				So(tt.conf.Redirect(tt.key, tt.ip), ShouldEqual, tt.exp)
			})
		}
	})
}

func TestParse(t *testing.T) {
	tests := []struct {
		act string
		Host
		err    error
		exp    string
		name   string
		reader Mapping
	}{
		{
			Host: Host{
				IP:     "127.0.0.1",
				Server: false,
			},
			act:    `{"badguys.com":{"IP":"0.0.0.0"}}`,
			err:    nil,
			exp:    "127.0.0.1",
			name:   "badguys.com",
			reader: Mapping{Contents: []byte(`address=/badguys.com/0.0.0.0`)},
		},
		{
			Host: Host{
				IP:     "127.0.0.1",
				Server: true,
			},
			act:    `{"xrated.com":{"IP":"0.0.0.0","Server":true}}`,
			err:    nil,
			exp:    "127.0.0.1",
			name:   "xrated.com",
			reader: Mapping{Contents: []byte(`server=/xrated.com/0.0.0.0`)},
		},
		{
			act:  `{}`,
			err:  errors.New("no dnsmasq configuration mapping entries found"),
			exp:  "127.0.0.1",
			name: "No dnsmasq entry",
			reader: Mapping{Contents: []byte(`# All files in this directory will be read by dnsmasq as 
# configuration files, except if their names end in 
# ".dpkg-dist",".dpkg-old" or ".dpkg-new"
#
# This can be changed by editing /etc/default/dnsmasq`)},
		},
	}
	Convey("Conf map should show each map entry", t, func() {
		c := make(Conf)
		for _, tt := range tests {
			Convey("current test: "+tt.name, func() {
				if err := c.Parse(&tt.reader); err != nil {
					So(err.Error(), ShouldEqual, tt.err.Error())
				}
				j, err := json.Marshal(c)
				So(err, ShouldBeNil)
				So(string(j), ShouldEqual, tt.act)
			})
		}
	})
}

func TestString(t *testing.T) {
	tests := []struct {
		conf Conf
		exp  string
	}{
		{
			conf: Conf{"badguys.com": Host{IP: "0.0.0.0", Server: false}},
			exp:  `{"badguys.com":{"IP":"0.0.0.0"}}`,
		},
		{
			exp: `null`,
		},
	}

	Convey("Testing String()", t, func() {
		for _, tt := range tests {
			So(tt.conf.String(), ShouldEqual, tt.exp)
		}
	})
}
