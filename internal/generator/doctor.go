package generator

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// DoctorCheck is one diagnostic result from RunDoctor.
type DoctorCheck struct {
	Name     string
	OK       bool
	Required bool // a failing required check is a real blocker, not advisory
	Detail   string
}

// RunDoctor runs the environment diagnostics genitz depends on, so a user
// sees what's missing up front instead of hitting a confusing failure
// mid-InstallStep. go is required (same binary CheckBinary gates on);
// git/docker/gh are advisory since genitz never requires them itself —
// Docker files are only generated, never run, and gh is only suggested.
func RunDoctor() []DoctorCheck {
	return []DoctorCheck{
		checkVersionedBinary("go", true, "go", "version"),
		checkVersionedBinary("git", false, "git", "--version"),
		checkVersionedBinary("docker", false, "docker", "--version"),
		checkVersionedBinary("gh", false, "gh", "--version"),
		checkGOPROXY(),
		checkNetwork(),
	}
}

func checkVersionedBinary(name string, required bool, cmd string, args ...string) DoctorCheck {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return DoctorCheck{Name: name, OK: false, Required: required, Detail: "not found on PATH"}
	}
	return DoctorCheck{Name: name, OK: true, Required: required, Detail: strings.TrimSpace(string(out))}
}

func checkGOPROXY() DoctorCheck {
	out, err := exec.Command("go", "env", "GOPROXY").Output()
	if err != nil {
		return DoctorCheck{Name: "GOPROXY", Detail: "could not read (go env failed)"}
	}
	proxy := strings.TrimSpace(string(out))
	return DoctorCheck{Name: "GOPROXY", OK: proxy != "off", Detail: proxy}
}

func checkNetwork() DoctorCheck {
	conn, err := net.DialTimeout("tcp", "proxy.golang.org:443", 2*time.Second)
	if err != nil {
		return DoctorCheck{Name: "network (proxy.golang.org)", Detail: "unreachable — go get for new deps will fail"}
	}
	_ = conn.Close()
	return DoctorCheck{Name: "network (proxy.golang.org)", OK: true, Detail: "reachable"}
}

// PrintDoctor renders RunDoctor's results — a check mark per line, with
// required failures called out distinctly from advisory ones.
func PrintDoctor(checks []DoctorCheck) {
	fmt.Println("🩺 Environment check:")
	anyRequiredFail := false
	for _, c := range checks {
		mark := "✔"
		if !c.OK {
			mark = "✖"
			if c.Required {
				anyRequiredFail = true
			}
		}
		fmt.Printf("   %s %-28s %s\n", mark, c.Name, c.Detail)
	}
	if anyRequiredFail {
		fmt.Println("\n⚠️  a required check failed — genitz needs it to work")
	}
}
