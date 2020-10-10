package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	flag "github.com/spf13/pflag"

	"golang.org/x/crypto/ssh"
)

// SSH global parameters
var (
	SSHPort     int    = 22
	SSHPassword string = ""
)

// NewConn establish ssh connection to iDRAC
func newConn(host, user, password string) *ssh.Client {
	// To be changed with an elegant way
	SSHPassword = password

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
			ssh.KeyboardInteractive(sshInteractive),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, SSHPort), config)
	if err != nil {
		panic(err)
	}
	return client
}

func sshInteractive(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
	answers = make([]string, len(questions))
	for n := range questions {
		answers[n] = SSHPassword
	}

	return answers, nil
}

func run(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		return "", err
	}
	return b.String(), nil
}

func extractInventory(rawInventory string) map[string]map[string]string {
	hwinventory := map[string]map[string]string{}

	lines := strings.Split(rawInventory, "\n")
	section := ""
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "-") {
			continue
		}

		if strings.HasPrefix(line, "[InstanceID:") {
			// Since section is not unique, a UUID is needed to distinguish section
			id, err := uuid.NewUUID()
			if err != nil {
				panic(err)
			}
			section = id.String() + "." + strings.Split(strings.Trim(strings.Trim(line, "["), "]"), " ")[1]
			hwinventory[section] = map[string]string{}
		} else {
			option := strings.Split(line, "=")
			k := strings.TrimSpace(option[0])
			v := strings.TrimSpace(option[1])
			hwinventory[section][k] = v
		}
	}

	return hwinventory
}

func sliceHas(s []string, e string) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func main() {
	dtypes := []string{}
	fields := []string{}
	var idrac, user, password string
	flag.StringSliceVarP(&dtypes, "types", "t", []string{"NIC", "FC"}, `Device types, "all" for all device types`)
	flag.StringSliceVarP(&fields, "fields", "f", []string{"all"}, `Fields to show, "all" for all fields`)
	flag.StringVarP(&idrac, "idrac", "i", "", "iDRAC FQDN/IP")
	flag.StringVarP(&user, "user", "u", "root", "iDRAC user")
	flag.StringVarP(&password, "password", "p", "calvin", "iDRAC password")
	flag.Parse()

	client := newConn(idrac, user, password)
	output, err := run(client, "racadm hwinventory")
	if err != nil {
		panic(err)
	}

	hwinventory := extractInventory(output)
	if hwinventory == nil || len(hwinventory) == 0 {
		panic(fmt.Sprintf("Fail to extract hardware inventory from output: %s", output))
	}

	target := map[string]map[string]string{}
	for k, v := range hwinventory {
		section := ""
		if sliceHas(dtypes, "all") {
			section = k
		} else {
			if sliceHas(dtypes, hwinventory[k]["Device Type"]) {
				section = k
			}
		}
		if section != "" {
			target[section] = map[string]string{}

			if sliceHas(fields, "all") {
				target[section] = v
			} else {
				for ik, iv := range v {
					if sliceHas(fields, ik) {
						target[section][ik] = iv
					}
				}
			}
		}
	}

	prettyInventory, _ := json.MarshalIndent(target, "", "  ")
	fmt.Printf("%s\n", prettyInventory)
}
