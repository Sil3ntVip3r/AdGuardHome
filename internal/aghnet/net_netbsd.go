//go:build netbsd

package aghnet

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/AdguardTeam/AdGuardHome/internal/aghos"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/osutil/executil"
)

func ifaceHasStaticIP(
	_ context.Context,
	_ executil.CommandConstructor,
	ifaceName string,
) (ok bool, err error) {
	const rcConfFilename = "etc/rc.conf"

	n := interfaceName(ifaceName)
	var isConfigured bool
	walker := aghos.FileWalker(func(r io.Reader) (_ []string, cont bool, err error) {
		isConfigured, ok, err = n.rcConfStaticConfig(r)

		return nil, true, err
	})
	_, err = walker.Walk(rootDirFS, rcConfFilename)
	if err != nil || isConfigured {
		return ok, err
	}

	filename := fmt.Sprintf("etc/ifconfig.%s", n)
	walker = aghos.FileWalker(ifconfigStaticConfig)

	return walker.Walk(rootDirFS, filename)
}

// rcConfStaticConfig checks if the interface is configured by /etc/rc.conf to
// have a static IP.
func (n interfaceName) rcConfStaticConfig(r io.Reader) (isConfigured, isStatic bool, err error) {
	s := bufio.NewScanner(r)
	pref := fmt.Sprintf("ifconfig_%s=", n)
	var config string
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if !strings.HasPrefix(line, pref) {
			continue
		}

		config = strings.Trim(strings.TrimSpace(line[len(pref):]), `"'`)
	}
	if err = s.Err(); err != nil || config == "" {
		return false, false, err
	}

	// NetBSD processes semicolon-delimited variable values as separate lines.
	isStatic, err = hasStaticIPv4Config(strings.NewReader(strings.ReplaceAll(config, ";", "\n")))

	return true, isStatic, err
}

// ifconfigStaticConfig checks if the interface is configured by an
// /etc/ifconfig.* file to have a static IP.
func ifconfigStaticConfig(r io.Reader) (_ []string, cont bool, err error) {
	var isStatic bool
	isStatic, err = hasStaticIPv4Config(r)

	return nil, !isStatic, err
}

// hasStaticIPv4Config reports whether the interface configuration contains an
// IPv4 address.
func hasStaticIPv4Config(r io.Reader) (isStatic bool, err error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		fields := strings.Fields(s.Text())
		switch {
		case
			len(fields) < 2,
			!strings.EqualFold(fields[0], "inet"),
			!netutil.IsValidIPString(fields[1]):
			continue
		default:
			return true, nil
		}
	}

	return false, s.Err()
}

func ifaceSetStaticIP(
	_ context.Context,
	_ *slog.Logger,
	_ executil.CommandConstructor,
	_ string,
) (err error) {
	return aghos.Unsupported("setting static ip")
}
