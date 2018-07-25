package credhub

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Set adds a credential in Credhub.
func (c *Client) Set(credential Credential, mode OverwriteMode, additionalPermissions []Permission) (Credential, error) {
	reqBody := struct {
		Credential
		Mode                  OverwriteMode `json:"mode"`
		AdditionalPermissions []Permission  `json:"additional_permissions,omitempty"`
	}{
		Credential: credential,
		Mode:       mode,
		AdditionalPermissions: additionalPermissions,
	}
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return Credential{}, err
	}

	var req *http.Request
	req, err = http.NewRequest("PUT", c.url+"/api/v1/data", bytes.NewBuffer(buf))
	if err != nil {
		return Credential{}, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return Credential{}, err
	}

	var cred Credential
	unmarshaller := json.NewDecoder(resp.Body)
	err = unmarshaller.Decode(&cred)

	return cred, err
}
