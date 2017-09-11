/*
   kube-secrets
   Copyright 2017 Jolene Engo <dev.toaster@gmail.com>

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"fmt"
	"flag"
	"sort"
	"strings"
	"io/ioutil"
	"encoding/base64"
	"github.com/go-yaml/yaml"
	"path/filepath"
)

const S_ERROR_FAILED_TO_PARSE      = "Failed to parse yml"
const S_ERROR_LOADING_SECRETS_FILE = "Error loading secrets file"
const S_ERROR_NOT_SECRETS_FILE     = "Not a Kubernetes secret"
const S_ERROR_OPENING_UPDATE_FILE  = "Failed opening file to update"
const S_ERROR_OPENING_EDITOR       = "Failed to launch editor"
const S_ERROR_PARAMETER_U          = "Can not use -u and -U together"
const S_FILE_CREATED               = "File created"
const S_FILE_UPDATED               = "File updated"
const S_KEY_DELETED                = "Key deleted"
const S_KEY_NOT_FOUND              = "Key not found"
const S_MISSING_EDITOR_VAR         = "Missing EDITOR environment variable"
const S_MISSING_PARAMETER_KEY      = "Missing require parameter key"
const S_NO_UPDATES                 = "No updates"

type ErrorMessage struct {
	Error string
}

type secrets struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Type       string `yaml:"type"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	}  `yaml:"metadata"`
	Data map[string]string `yaml:"data"`
	updateValue string
}

// Short cut handler
func checkError(e error) {
	if e != nil {
		fmt.Println(e.Error())
		os.Exit(1)
	}
}

func (c *secrets) loadSecretsFile(filename string) error {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.New(S_ERROR_LOADING_SECRETS_FILE)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return errors.New(S_ERROR_FAILED_TO_PARSE)
	}

	// Obviously we can only work with secrets files
	if c.Kind != "Secret" {
		return errors.New(S_ERROR_NOT_SECRETS_FILE)
	}

	return nil
}

func listKeys(filename string) (error, string) {
	var s secrets
	var buffer bytes.Buffer	
	
	err := s.loadSecretsFile(filename)
	if err != nil {
		return errors.New(S_ERROR_LOADING_SECRETS_FILE), ""
	}

	// Sort to make it easier to unit test
	keys := make([]string, 0, len(s.Data))
	for key := range s.Data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		buffer.WriteString(fmt.Sprintln(key))		
	}

	return nil, buffer.String()
}

func (s *secrets) hasKey(key string) bool {
	_, result := s.Data[key]

	return result
}

func (s *secrets) checkForUpdateValue(updateString string, updateFile string) error {
	if updateString != "" {
		s.updateValue = updateString
	} else if updateFile != "" {
		data, err := ioutil.ReadFile(updateFile)
		if err != nil {
			err := errors.New(S_ERROR_OPENING_UPDATE_FILE)
			return err
		}
		s.updateValue = string(data)
	}

	return nil
}

func editor(openWith []byte) (string, error) {
	editor := os.Getenv("EDITOR")

	if editor == "" {
		return "", errors.New(S_MISSING_EDITOR_VAR)
	}

	tmpDir := os.TempDir()
	tmpFile, tmpFileErr := ioutil.TempFile(tmpDir, "tmp-kube-secrets.")
	if tmpFileErr != nil {
		fmt.Printf("Error %s while creating tempFile", tmpFileErr)
	}
	
	path, err := exec.LookPath(editor)
	if err != nil {
		return "", errors.New(S_ERROR_OPENING_EDITOR)
	}

	if _, err := tmpFile.Write(openWith); err != nil {
		fmt.Println("Failed to write to temp file")
		os.Exit(3)
	}

	cmd        := exec.Command(path, tmpFile.Name())
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	err        = cmd.Start()
	if err != nil {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	tmpFile.Close()
	os.Remove(tmpFile.Name())

	// VI / VIM very annoyingly adds a newline to all saved files
	// Which leads to problems with values like usernames and passwords
	if editor == "vi" || editor == "vim" {
		b = []byte(strings.TrimSuffix(string(b), "\n"))
	}

	result := base64.StdEncoding.EncodeToString(b)

	return result, nil
}

func createNewSecrets(filename string, key string, updateString string, updateFile string) error {
	s            := secrets{}
	s.APIVersion = "v1"
	s.Kind       = "Secret"
	s.Type       = "Opaque"

	s.checkForUpdateValue(updateString, updateFile)

	ext := filepath.Ext(filename)
	// The secret name will have the directory and ext stripped
	// /tmp/test.yml becomes test
	s.Metadata.Name = filepath.Base(filename[0:len(filename)-len(ext)])

	updateValue := ""

	if s.updateValue == "" {
		empty   := []byte("")
		uv, err := editor(empty)
		updateValue = uv
		if err != nil {
			return err
		}
	} else {
		updateValue = base64.StdEncoding.EncodeToString([]byte(s.updateValue))
	}

	s.Data      = make(map[string]string)
	s.Data[key] = updateValue
	s.writeSecrets(filename)

	return nil
}

func deleteSecret(filename string, key string) error {
	var s secrets
	
	load_err := s.loadSecretsFile(filename)
	if load_err != nil {
		return load_err
	}

	err := s.validateKey(key)
	if err != nil {
		return err
	}

	delete(s.Data, key)
	s.writeSecrets(filename)

	return nil
}

func showSecretsFile(filename string, key string) (error, string) {
	var s secrets

	err := s.loadSecretsFile(filename)
	if err != nil {
		return err, ""
	}

	err = s.validateKey(key)
	if err != nil {
		return err, ""
	}

	sDec, _ := base64.StdEncoding.DecodeString(s.Data[key])

	return nil, string(sDec)
}

func updateSecretsFile(filename string, key string, updateString string, updateFile string) (error, string) {
	var s secrets
	
	err := s.loadSecretsFile(filename)
	var newValue string

	if err != nil {
		return errors.New(S_ERROR_LOADING_SECRETS_FILE), ""
	}

	s.checkForUpdateValue(updateString, updateFile)	

	if s.updateValue != "" {
		newValue = base64.StdEncoding.EncodeToString([]byte(s.updateValue))
	} else {
		sDec, err           := base64.StdEncoding.DecodeString(s.Data[key])			
		newEditorValue, err := editor(sDec)			
		if err != nil {
			return errors.New(S_ERROR_OPENING_EDITOR), ""
		}
		newValue = newEditorValue
	}

	if (newValue == s.Data[key]) {
		return nil, S_NO_UPDATES
	}
	
	s.Data[key] = newValue
	// TODO: Report if there was any changes saved
	s.writeSecrets(filename)

	return nil, S_FILE_UPDATED
}

func (s *secrets) validateKey(key string) error {
	if key == "" {
		return errors.New(S_MISSING_PARAMETER_KEY)
	}

	has_key := s.hasKey(key)
	
	if ! has_key {
		return errors.New(S_KEY_NOT_FOUND)
	}

	return nil
}

func (s *secrets) writeSecrets(filename string) error {
	// If no namespace is set, use default
	if s.Metadata.Namespace == "" {
		s.Metadata.Namespace = "default"
	}

	newYML, err := yaml.Marshal(&s)
	if err != nil {
		fmt.Printf("error: %v", err)
	}

	writeErr := ioutil.WriteFile(filename, newYML, 0644)

	return writeErr
}

func printUsage() string {
	var msg string

	msg = fmt.Sprintf("kube-secrets version %s\n\n", VERSION)
	msg = msg + "Usage: kube-secrets <options> <command> <file>\n"
	msg = msg + "\nAvailable commands:\n"
	msg = msg + "   create\t\tCreate new secret file\n"
	msg = msg + "   delete\t\tRemove key\n"
	msg = msg + "   help \t\tPrint usage\n"
	msg = msg + "   keys \t\tList all keys in secret file\n"
	msg = msg + "   update\t\tUpdate value of key or create new key\n"
	msg = msg + "   version\t\tPrint version\n\n"
	msg = msg + "Options:\n"
	msg = msg + "   -u   \t\tString to set value with\n"
	msg = msg + "   -U   \t\tFile to set value with\n"

	return msg
}

func parseArgs() (error, string) {
	updateStringPtr := flag.String("u", "", "STRING")
	updateFilePtr   := flag.String("U", "", "FILENAME")
	helpPtr1        := flag.Bool("h", false, "")
	helpPtr2        := flag.Bool("help", false, "")
	
	flag.Parse()

	command  := flag.Arg(0)
	filename := flag.Arg(1)
	key      := flag.Arg(2)

	if *helpPtr1 || *helpPtr2 {
		return nil, printUsage()
	}

	if *updateStringPtr != "" && *updateFilePtr != "" {
		return nil, S_ERROR_PARAMETER_U
	}
	
	switch command  {
		case "create":
			if key == "" {
				return errors.New(S_MISSING_PARAMETER_KEY), ""
			}
			err := createNewSecrets(filename, key, *updateStringPtr, *updateFilePtr)
			checkError(err)

			return nil, S_FILE_CREATED

		case "delete":
			err := deleteSecret(filename, key)			
			checkError(err)			

			return nil, S_KEY_DELETED

		case "keys":
			err, resultString := listKeys(filename)
			checkError(err)			
			
			return nil, resultString

		case "help":
			return nil, printUsage()
		
			return nil, ""

		case "show":
			if key == "" {
				return errors.New(S_MISSING_PARAMETER_KEY), ""
			}
			err, keyValue := showSecretsFile(filename, key)
			checkError(err)			
			
			return nil, keyValue

		case "update":
			err, updatedStatus := updateSecretsFile(filename, key, *updateStringPtr, *updateFilePtr)
			checkError(err)			
			
			return nil, updatedStatus
			
		case "version":
			fmt.Printf("kube-secrets version %s\n", VERSION)
			return nil, ""

		default:
			return nil, printUsage()
	}


	return nil, ""
}
