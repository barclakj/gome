package env

// Courtesy of https://github.com/ShowMax/go-fqdn/blob/master/fqdn.go

import (
	"log"
	"net"
	"os"
	"os/user"
	"strings"
)

/* Returns the origin in form <user>@<host> */
func GetOrigin() string {
	return GetUser() + "@" + GetHostname()
}

/* Returns the current user running the gome process */
func GetUser() string {
	usr, _ := user.Current()
	return usr.Username
}

/* Returns the current users home directory */
func GetHome() string {
	usr, _ := user.Current()
	return usr.HomeDir
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

/* Checks if the specified address if local (true) or not (false) */
func IsLocalAddress(testIP string) bool {
	local := false
	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.String() == testIP {
				log.Printf("Local IP...")
				local = true
				break
			}
			// process IP address
		}
		if local == true {
			break
		}
	}
	return local
}
