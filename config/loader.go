package config

import (
	"encoding/json"
	"fmt"
	"github.com/paroxity/portal/server"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type config struct {
	Query struct {
		MaxPlayers int    `json:"maxPlayers"`
		MOTD       string `json:"motd"`
	} `json:"query"`
	Whitelist struct {
		Enabled bool     `json:"enabled"`
		Players []string `json:"players"`
	} `json:"whitelist"`
	Proxy struct {
		BindAddress    string                       `json:"bindAddress"`
		Groups         map[string]map[string]string `json:"groups"`
		DefaultGroup   string                       `json:"defaultGroup"`
		Authentication bool                         `json:"authentication"`
		ResourcesDir   string                       `json:"resourcesDir"`
		ForceTextures  bool                         `json:"forceTextures"`
	} `json:"proxy"`
	Socket struct {
		BindAddress string `json:"bindAddress"`
		Secret      string `json:"secret"`
	} `json:"socket"`
	Logger struct {
		File  string `json:"file"`
		Debug bool   `json:"debug"`
	} `json:"logger"`
}

func init() {
	if err := Load(); err != nil {
		log.Printf("Unable to load config: %v\n", err)
	}
}

func Load() error {
	c := config{}
	c.Query.MOTD = "Portal"
	c.Proxy.BindAddress = "0.0.0.0:19132"
	c.Proxy.Groups = map[string]map[string]string{
		"Hub": {
			"Hub1": "127.0.0.1:19133",
		},
	}
	c.Proxy.DefaultGroup = "Hub"
	c.Proxy.Authentication = true
	c.Socket.BindAddress = "127.0.0.1:19131"
	c.Logger.File = "proxy.log"

	if _, err := os.Stat("config.json"); os.IsNotExist(err) {
		data, err := json.MarshalIndent(c, "", "    ")
		if err != nil {
			return fmt.Errorf("failed encoding default config: %v", err)
		}
		if err := ioutil.WriteFile("config.json", data, 0644); err != nil {
			return fmt.Errorf("failed creating config: %v", err)
		}
	} else {
		data, err := ioutil.ReadFile("config.json")
		if err != nil {
			return fmt.Errorf("error reading config: %v", err)
		}
		if err := json.Unmarshal(data, &c); err != nil {
			return fmt.Errorf("error decoding config: %v", err)
		}
	}

	maxPlayers = c.Query.MaxPlayers
	motd = c.Query.MOTD
	whitelist = c.Whitelist.Enabled
	whitelisted = c.Whitelist.Players

	if len(c.Proxy.Groups) == 0 {
		return fmt.Errorf("groups are empty")
	}

	for name, group := range c.Proxy.Groups {
		g := server.NewGroup(name)
		for name, address := range group {
			g.AddServer(server.New(name, g.Name(), address))
		}

		server.AddGroup(g)
	}

	g, ok := server.GroupFromName(c.Proxy.DefaultGroup)
	if !ok {
		return fmt.Errorf("default group %s not found", c.Proxy.DefaultGroup)
	}
	server.SetDefaultGroup(g)

	authentication = c.Proxy.Authentication
	bindAddress = c.Proxy.BindAddress
	files, err := ioutil.ReadDir(c.Proxy.ResourcesDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, file := range files {
		pack, err := resource.Compile(path.Join(c.Proxy.ResourcesDir, file.Name()))
		if err != nil {
			return err
		}
		resourcePacks = append(resourcePacks, pack)
	}
	forceTextures = c.Proxy.ForceTextures

	socketAddress = c.Socket.BindAddress
	socketSecret = c.Socket.Secret

	logFile = c.Logger.File
	logLevel = logrus.InfoLevel
	if c.Logger.Debug {
		logLevel = logrus.DebugLevel
	}

	return nil
}
