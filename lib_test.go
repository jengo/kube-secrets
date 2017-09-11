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
	"os"
	"io/ioutil"
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
)

func Test_hasKey(t *testing.T) {
	s              := secrets{}
	s.Data         = make(map[string]string)
	s.Data["TEST"] = "XXX"

	r1             := s.hasKey("TEST")
	assert.True(t, r1)

	r2             := s.hasKey("TEST2")
	assert.False(t, r2)
}

func Test_listKeys(t *testing.T) {
	err, result := listKeys("/tmp/test_data/sample.yml")

	assert.Nil(t, err)
	
	assert.Equal(t, "INVALID_BASE64\nMYSQL_ROOT_PASSWORD\naaa_data\ntest_delete\ntls.key\nupdate_same_value\nzzz_data\n", result)
}

func Test_listKeys_error_missing_file(t *testing.T) {
	err, result := listKeys("/tmp/test_data/MISSING_FILE.yml")

	assert.Equal(t, errors.New(S_ERROR_LOADING_SECRETS_FILE), err)
	assert.Equal(t, "", result)
}

func Test_listKeys_error_invalid_file(t *testing.T) {
	err, result := listKeys("/tmp/test_data/invalid.yml")

	assert.Equal(t, errors.New(S_ERROR_LOADING_SECRETS_FILE), err)
	assert.Equal(t, "", result)
}

func Test_showSecretsFile(t *testing.T) { 
	err, result := showSecretsFile("/tmp/test_data/sample.yml", "MYSQL_ROOT_PASSWORD")

	assert.Nil(t, err)
	assert.Equal(t, "my-super-secret-squirrel-password\n", result)
}

func Test_showSecretsFile_invalid_key(t *testing.T) { 
	err, result := showSecretsFile("/tmp/test_data/sample.yml", "MISSING_KEY")

	assert.Equal(t, errors.New(S_KEY_NOT_FOUND), err)
	assert.Equal(t, "", result)
}

func Test_showSecretsFile_error_missing_file(t *testing.T) {
	err, result := showSecretsFile("/tmp/test_data/MISSING_FILE.yml", "MYSQL_ROOT_PASSWORD")
	
	assert.Equal(t, errors.New(S_ERROR_LOADING_SECRETS_FILE), err)
	assert.Equal(t, "", result)
}

func Test_Test_showSecretsFile_error_invalid_file(t *testing.T) {
	err, result := showSecretsFile("/tmp/test_data/invalid.yml", "MYSQL_ROOT_PASSWORD")

	assert.Equal(t, errors.New(S_ERROR_FAILED_TO_PARSE), err)
	assert.Equal(t, "", result)
}

func Test_Test_showSecretsFile_error_not_secret(t *testing.T) {
	err, result := showSecretsFile("/tmp/test_data/configmap.yml", "MYSQL_ROOT_PASSWORD")

	assert.Equal(t, errors.New(S_ERROR_NOT_SECRETS_FILE), err)
	assert.Equal(t, "", result)
}

func Test_loadSecretsFile(t *testing.T) {
	var s secrets
	err := s.loadSecretsFile("/tmp/test_data/sample.yml")

	assert.Nil(t, err)
}

func Test_loadSecretsFile_invalid(t *testing.T) {
	var s secrets
	err := s.loadSecretsFile("/tmp/test_data/invalid.yml")

	assert.Equal(t, errors.New(S_ERROR_FAILED_TO_PARSE), err)
}

func Test_loadSecretsFile_not_secret(t *testing.T) {
	var s secrets
	err := s.loadSecretsFile("/tmp/test_data/configmap.yml")

	assert.Equal(t, errors.New(S_ERROR_NOT_SECRETS_FILE), err)
}

func Test_validateKey(t *testing.T) {
	var s secrets
	s.loadSecretsFile("/tmp/test_data/sample.yml")
	err := s.validateKey("MYSQL_ROOT_PASSWORD")

	assert.Nil(t, err)
}

func Test_validateKey_not_found(t *testing.T) {
	var s secrets
	s.loadSecretsFile("/tmp/test_data/sample.yml")
	err := s.validateKey("MYSQL_ROOT_PASSWORD_INVALID")

	assert.Equal(t, errors.New(S_KEY_NOT_FOUND), err)
}

func Test_validateKey_missing_parameter(t *testing.T) {
	var s secrets
	s.loadSecretsFile("/tmp/test_data/sample.yml")
	err := s.validateKey("")

	assert.Equal(t, errors.New(S_MISSING_PARAMETER_KEY), err)
}

func Test_deleteSecret(t *testing.T) {
	err := deleteSecret("/tmp/test_data/sample.yml", "test_delete")

	assert.Nil(t, err)
}

func Test_deleteSecret_missing(t *testing.T) {
	err := deleteSecret("/tmp/test_data/sample.yml", "NOT_REAL_KEY")

	assert.Equal(t, errors.New(S_KEY_NOT_FOUND), err)
}

func Test_deleteSecret_invalid_file(t *testing.T) {
	err := deleteSecret("/tmp/test_data/MISSING_FILE.yml", "NOT_REAL_KEY")

	assert.Equal(t, errors.New(S_ERROR_LOADING_SECRETS_FILE), err)
}

func Test_createNewSecrets_by_value(t *testing.T) {
	err := createNewSecrets("/tmp/test-create.yml", "test", "test_val123", "")
	assert.Nil(t, err)
	
	new, new_err := ioutil.ReadFile("/tmp/test-create.yml")
	assert.Nil(t, new_err)

	expected, expected_err := ioutil.ReadFile("/tmp/test_data/test-create.yml")
	assert.Nil(t, expected_err)

	assert.Equal(t, string(expected), string(new))
}

func Test_createNewSecrets_by_file(t *testing.T) {
	err := createNewSecrets("/tmp/test-create.yml", "test", "", "/tmp/test_data/test-file-data.txt")
	assert.Nil(t, err)
	
	new, new_err := ioutil.ReadFile("/tmp/test-create.yml")
	assert.Nil(t, new_err)

	expected, expected_err := ioutil.ReadFile("/tmp/test_data/test-create.yml")
	assert.Nil(t, expected_err)

	assert.Equal(t, string(expected), string(new))
}

func Test_createNewSecrets_by_editor(t *testing.T) {
	os.Setenv("EDITOR", "test_data/editor_emulate.sh")
	
	err := createNewSecrets("/tmp/test-create-by-editor.yml", "test", "", "")
	assert.Nil(t, err)
	
	new, new_err := ioutil.ReadFile("/tmp/test-create-by-editor.yml")
	assert.Nil(t, new_err)

	expected, expected_err := ioutil.ReadFile("/tmp/test_data/test-create-by-editor.yml")
	assert.Nil(t, expected_err)

	assert.Equal(t, string(expected), string(new))
}

func Test_editor(t *testing.T) {
	os.Setenv("EDITOR", "test_data/editor_emulate.sh")

	empty   := []byte("")

	result, err := editor(empty)

	assert.Nil(t, err)
	assert.Equal(t, "VEVTVF9EQVRBCg==", result)
}

func Test_editor_no_environment_variable(t *testing.T) {
	os.Setenv("EDITOR", "")

	empty   := []byte("")

	result, err := editor(empty)

	assert.Equal(t, errors.New(S_MISSING_EDITOR_VAR), err)
	assert.Equal(t, "", result)
}

func Test_checkForUpdateValue_can_not_open(t *testing.T) {
	var s secrets

	err := s.checkForUpdateValue("", "/tmp/test_can_not_open.yml")
	assert.Equal(t, errors.New(S_ERROR_OPENING_UPDATE_FILE), err)
}

func Test_updateSecretsFile_error_loading(t *testing.T) {
	err, value := updateSecretsFile("/tmp/test_data/no_such_file.yml", "update_error", "", "")

	assert.Equal(t, errors.New(S_ERROR_LOADING_SECRETS_FILE), err)
	assert.Equal(t, "", value)
}

func Test_updateSecretsFile_error_opening_editor(t *testing.T) {
	os.Setenv("EDITOR", "no_such_editor")
	err, value := updateSecretsFile("/tmp/test_data/sample.yml", "update_error", "", "")

	assert.Equal(t, errors.New(S_ERROR_OPENING_EDITOR), err)
	assert.Equal(t, "", value)
}

func Test_updateSecretsFile_update_same_value_string(t *testing.T) {
	err, value := updateSecretsFile("/tmp/test_data/sample.yml", "update_same_value", "test_val123", "")

	assert.Nil(t, err)
	assert.Equal(t, S_NO_UPDATES, value)
}

func Test_updateSecretsFile_update_same_value_file(t *testing.T) {
	err, value := updateSecretsFile("/tmp/test_data/sample.yml", "update_same_value", "", "/tmp/test_data/test-file-data.txt")

	assert.Nil(t, err)
	assert.Equal(t, S_NO_UPDATES, value)
}

func Test_updateSecretsFile_update_new_key(t *testing.T) {
	err, value := updateSecretsFile("/tmp/test_data/sample.yml", "newkey", "newvalue", "")

	assert.Nil(t, err)
	assert.Equal(t, S_FILE_UPDATED, value)

	// Now open that same file to test the value
	var s secrets
	s.loadSecretsFile("/tmp/test_data/sample.yml")
	// The value should be base64
	assert.Equal(t, s.Data["newkey"], "bmV3dmFsdWU=")
}

// Honestly, totally lame test
func Test_printUsage(t *testing.T) {
	result := printUsage()

	assert.NotEqual(t, result, "")
}

