package env

// Courtesy of https://github.com/ShowMax/go-fqdn/blob/master/fqdn.go

import (
	"log"
	"net"
	"os"
	"os/user"
	"strings"
)

func GetOrigin() string {
	return GetUser() + "@" + GetHostname()
}

func GetUser() string {
	usr, _ := user.Current()
	return usr.Username
}

// Get Fully Qualified Domain Name
// returns "unknown" or hostanme in case of error
func GetHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname
			}
			fqdn := hosts[0]
			log.Printf("Found hostname: %s\n", fqdn)

			fqdn = strings.TrimSuffix(fqdn, ".") // return fqdn without trailing dot
			return fqdn
		}
	}
	return hostname
}
