package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/user"
	"strings"

	"github.com/coreos/pkg/capnslog"
	"github.com/ids/clairctl/xstrings"
	"github.com/ids/xnet"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

var log = capnslog.NewPackageLogger("github.com/ids/clairctl", "config")
var serverPort = 0

var errNoInterfaceProvided = errors.New("could not load configuration: no interface provided")
var errInvalidInterface = errors.New("Interface does not exist")
var ErrLoginNotFound = errors.New("user is not log in")

var IsLocal = true
var Insecure = false
var NoClean = false

var ImageName string
var ClusterImages map[string]struct{}

type reportConfig struct {
	Path, Format string
}
type clairConfig struct {
	URI              string
	Port, HealthPort int
	Report           reportConfig
}
type authConfig struct {
	InsecureSkipVerify bool
}
type clairctlConfig struct {
	IP, Interface, TempFolder string
	Port                      int
}
type docker struct {
	InsecureRegistries []string
}

type config struct {
	Clair    clairConfig
	Auth     authConfig
	Clairctl clairctlConfig
	Docker   docker
}

// Init reads in config file and ENV variables if set.
func Init(cfgFile string, logLevel string, noClean bool) {

	NoClean = noClean

	lvl := capnslog.WARNING
	if logLevel != "" {
		// Initialize logging system
		var err error
		lvl, err = capnslog.ParseLevel(strings.ToUpper(logLevel))
		if err != nil {
			log.Warningf("Wrong Log level %v, defaults to [Warning]", logLevel)
			lvl = capnslog.WARNING
		}

	}
	capnslog.SetGlobalLogLevel(lvl)
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(os.Stdout, false))

	viper.SetEnvPrefix("clairctl")
	viper.SetConfigName("clairctl")        // name of config file (without extension)
	viper.AddConfigPath("$HOME/.clairctl") // adding home directory as first search path
	viper.AddConfigPath(".")               // adding home directory as first search path
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	err := viper.ReadInConfig()
	if err != nil {
		e, ok := err.(viper.ConfigParseError)
		if ok {
			log.Fatalf("error parsing config file: %v", e)
		}
		log.Debugf("No config file used")
	} else {
		log.Debugf("Using config file: %v", viper.ConfigFileUsed())
	}

	if viper.Get("clair.uri") == nil {
		viper.Set("clair.uri", "http://localhost")
	}
	if viper.Get("clair.port") == nil {
		viper.Set("clair.port", "6060")
	}
	if viper.Get("clair.healthPort") == nil {
		viper.Set("clair.healthPort", "6061")
	}

	if viper.Get("clair.report.path") == nil {
		viper.Set("clair.report.path", "reports")
	}
	if viper.Get("clair.report.format") == nil {
		viper.Set("clair.report.format", "html")
	}
	if viper.Get("auth.insecureSkipVerify") == nil {
		viper.Set("auth.insecureSkipVerify", "true")
	}
	if viper.Get("clairctl.ip") == nil {
		viper.Set("clairctl.ip", "")
	}
	if viper.Get("clairctl.port") == nil {
		viper.Set("clairctl.port", 0)
	}
	if viper.Get("clairctl.interface") == nil {
		viper.Set("clairctl.interface", "")
	}
	if viper.Get("clairctl.tempFolder") == nil {
		viper.Set("clairctl.tempFolder", "/tmp/clairctl")
	}
	if viper.Get("notifier.severity") == nil {
		viper.Set("notifier.severity", "Critical")
	}

}

func TmpLocal() string {
	return viper.GetString("clairctl.tempFolder")
}

func values() config {
	return config{
		Clair: clairConfig{
			URI:        viper.GetString("clair.uri"),
			Port:       viper.GetInt("clair.port"),
			HealthPort: viper.GetInt("clair.healthPort"),
			Report: reportConfig{
				Path:   viper.GetString("clair.report.path"),
				Format: viper.GetString("clair.report.format"),
			},
		},
		Auth: authConfig{
			InsecureSkipVerify: viper.GetBool("auth.insecureSkipVerify"),
		},
		Clairctl: clairctlConfig{
			IP:         viper.GetString("clairctl.ip"),
			Port:       viper.GetInt("clairctl.port"),
			TempFolder: viper.GetString("clairctl.tempFolder"),
			Interface:  viper.GetString("clairctl.interface"),
		},
		Docker: docker{
			InsecureRegistries: viper.GetStringSlice("docker.insecure-registries"),
		},
	}
}

func Print() {
	cfg := values()
	cfgBytes, err := yaml.Marshal(cfg)
	if err != nil {
		log.Fatalf("marshalling configuration: %v", err)
	}

	fmt.Println("Configuration")
	fmt.Printf("%v", string(cfgBytes))
}

func ClairctlHome() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	p := usr.HomeDir + "/.clairctl"

	if _, err := os.Stat(p); os.IsNotExist(err) {
		os.Mkdir(p, 0700)
	}
	return p
}

type Login struct {
	Username string
	Password string
}

type loginMapping map[string]Login

func ClairctlConfig() string {
	return ClairctlHome() + "/config.json"
}

func AddLogin(registry string, login Login) error {
	var logins loginMapping

	if err := readConfigFile(&logins, ClairctlConfig()); err != nil {
		return fmt.Errorf("reading clairctl file: %v", err)
	}

	logins[registry] = login

	if err := writeConfigFile(logins, ClairctlConfig()); err != nil {
		return fmt.Errorf("indenting login: %v", err)
	}

	return nil
}
func GetLogin(registry string) (Login, error) {
	if _, err := os.Stat(ClairctlConfig()); err == nil {
		var logins loginMapping

		if err := readConfigFile(&logins, ClairctlConfig()); err != nil {
			return Login{}, fmt.Errorf("reading clairctl file: %v", err)
		}

		if login, present := logins[registry]; present {
			d, err := base64.StdEncoding.DecodeString(login.Password)
			if err != nil {
				return Login{}, fmt.Errorf("decoding password: %v", err)
			}
			login.Password = string(d)
			return login, nil
		}
	}
	return Login{}, ErrLoginNotFound
}

func RemoveLogin(registry string) (bool, error) {
	if _, err := os.Stat(ClairctlConfig()); err == nil {
		var logins loginMapping

		if err := readConfigFile(&logins, ClairctlConfig()); err != nil {
			return false, fmt.Errorf("reading clairctl file: %v", err)
		}

		if _, present := logins[registry]; present {
			delete(logins, registry)

			if err := writeConfigFile(logins, ClairctlConfig()); err != nil {
				return false, fmt.Errorf("indenting login: %v", err)
			}

			return true, nil
		}
	}
	return false, nil
}

func readConfigFile(logins *loginMapping, file string) error {
	if _, err := os.Stat(file); err == nil {
		f, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(f, &logins); err != nil {
			return err
		}
	} else {
		*logins = loginMapping{}
	}
	return nil
}

func writeConfigFile(logins loginMapping, file string) error {
	s, err := xstrings.ToIndentJSON(logins)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(file, s, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func getFreePort(localIP string) (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:0", localIP))
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	err = l.Close()
	return port, nil
}

//LocalServerIP return the local clairctl server IP
func LocalServerIP() (string, error) {
	localIP := viper.GetString("clairctl.ip")
	localInterfaceConfig := viper.GetString("clairctl.interface")

	//localPort := viper.GetString("clairctl.port")
	if serverPort == 0 {
		port, err := getFreePort(localIP)
		if err != nil {
			log.Fatal(err)
		}

		log.Debugf("port %v is free", port)
		serverPort = port
	}
	localPort := fmt.Sprintf("%v", serverPort)

	if localIP == "" {
		log.Info("retrieving interface for local IP")
		var err error
		var localInterface net.Interface
		localInterface, err = translateInterface(localInterfaceConfig)
		if err != nil {
			return "", fmt.Errorf("retrieving interface: %v", err)
		}

		localIP, err = xnet.IPv4(localInterface)
		if err != nil {
			return "", fmt.Errorf("retrieving interface ip: %v", err)
		}
	}
	return strings.TrimSpace(localIP) + ":" + localPort, nil
}

func translateInterface(localInterface string) (net.Interface, error) {

	if localInterface != "" {
		log.Debug("interface provided, looking for " + localInterface)
		netInterface, err := net.InterfaceByName(localInterface)
		if err != nil {
			return net.Interface{}, err
		}
		return *netInterface, nil
	}

	log.Debug("no interface provided, looking for docker0")
	netInterface, err := net.InterfaceByName("docker0")
	if err != nil {
		log.Debug("docker0 not found, looking for first connected broadcast interface")
		interfaces, err := net.Interfaces()
		if err != nil {
			return net.Interface{}, err
		}

		i, err := xnet.First(xnet.Filter(interfaces, xnet.IsBroadcast), xnet.HasAddr)
		if err != nil {
			return net.Interface{}, err
		}
		return i, nil
	}

	return *netInterface, nil

}

func Clean() error {
	if IsLocal && !NoClean {
		log.Debug("cleaning temporary local repository")
		err := os.RemoveAll(TmpLocal())

		if err != nil {
			return fmt.Errorf("cleaning temporary local repository: %v", err)
		}
	}

	return nil
}
