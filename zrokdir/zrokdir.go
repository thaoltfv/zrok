package zrokdir

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

type Environment struct {
	Token       string `json:"zrok_token"`
	ZId         string `json:"ziti_identity"`
	ApiEndpoint string `json:"api_endpoint"`
}

func LoadEnvironment() (*Environment, error) {
	ef, err := environmentFile()
	if err != nil {
		return nil, errors.Wrap(err, "error getting environment file")
	}
	data, err := os.ReadFile(ef)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading environment file '%v'", ef)
	}
	env := &Environment{}
	if err := json.Unmarshal(data, env); err != nil {
		return nil, errors.Wrapf(err, "error unmarshaling environment file '%v'", ef)
	}
	return env, nil
}

func SaveEnvironment(env *Environment) error {
	data, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		return errors.Wrap(err, "error marshaling environment")
	}
	ef, err := environmentFile()
	if err != nil {
		return errors.Wrap(err, "error getting environment file")
	}
	if err := os.MkdirAll(filepath.Dir(ef), os.FileMode(0700)); err != nil {
		return errors.Wrapf(err, "error creating zrokdir path '%v'", filepath.Dir(ef))
	}
	if err := os.WriteFile(ef, data, os.FileMode(0600)); err != nil {
		return errors.Wrap(err, "error saving environment file")
	}
	return nil
}

func DeleteEnvironment() error {
	ef, err := environmentFile()
	if err != nil {
		return errors.Wrap(err, "error getting environment file")
	}
	if err := os.Remove(ef); err != nil {
		return errors.Wrap(err, "error removing environment file")
	}

	return nil
}

func ZitiIdentityFile(name string) (string, error) {
	zrd, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "identities", fmt.Sprintf("%v.json", name)), nil
}

func SaveZitiIdentity(name, data string) error {
	zif, err := ZitiIdentityFile(name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(zif), os.FileMode(0700)); err != nil {
		return errors.Wrapf(err, "error creating zrokdir path '%v'", filepath.Dir(zif))
	}
	if err := os.WriteFile(zif, []byte(data), os.FileMode(0600)); err != nil {
		return errors.Wrapf(err, "error writing ziti identity file '%v'", zif)
	}
	return nil
}

func DeleteZitiIdentity(name string) error {
	zif, err := ZitiIdentityFile(name)
	if err != nil {
		return errors.Wrapf(err, "error getting ziti identity file path for '%v'", name)
	}
	if err := os.Remove(zif); err != nil {
		return errors.Wrapf(err, "error removing ziti identity file '%v'", zif)
	}
	return nil
}

func environmentFile() (string, error) {
	zrd, err := zrokDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(zrd, "environment.json"), nil
}

func zrokDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zrok"), nil
}
