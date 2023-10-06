package provider

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	ApiUrl = "https://api.deno.com/v1"
)

var (
	//nolint:unused
	organizationId = os.Getenv("DEPLOY_ORGANIZATION_ID")
	//nolint:unused
	token = os.Getenv("DEPLOY_ORGANIZATION_ACCESS_TOKEN")
	//nolint:unused
	bearer = "Bearer " + token
)

// api types
// OrganizationResponse is returned by `GET /organizations/<organizationId>`.
type OrganizationResponse struct {
	Id        string `json:"id"` // todo: uuid
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"` // todo: DateTime<Utc>
	UpdatedAt string `json:"updatedAt"` // todo: DateTime<Utc>
}

// ProjectResponse is an element of the array that is returned by
// `GET /organizations/<organizationId>/projects` and by a successful
// create via `POST /organizations/<organizationId/projects`.
type ProjectResponse struct {
	Id        string `json:"id"` // todo: uuid
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"` // todo: DateTime<Utc>
	UpdatedAt string `json:"updatedAt"` // todo: DateTime<Utc>
}

// ApiCreateProjectRequest is an optional payload for
// `POST /organizations/<organizationId/projects`. If a name
// is not provided, the API creates a random name.
type ApiCreateProjectRequest struct {
	Name string `json:"name"`
}

// ApiUpdateProjectRequest is the payload to update project details
// via `PATCH /projects/<projectId>`.
type ApiUpdateProjectRequest struct {
	Name string `json:"name"`
}

// ApiDeploymentResponse is the response to
// `POST /projects/${projectId}/deployments`. Note that creating a deployment
// is asynchronous with this response (the `Status` field will be `pending`).
// Poll `GET /deployments/<deploymentId>` (which returns a version of this same
// type) for the final status.
type ApiDeploymentResponse struct {
}

//nolint:unused
func main() {
	log := log.Default()

	client := &http.Client{
		Timeout: time.Duration(15) * time.Second,
	}

	getOrganizationById(client, log)
	getProjectsList(client, log)
	createAndDeleteProject(client, log)
}

//nolint:unused
func invokeGet(client *http.Client, log *log.Logger, url string) *http.Response {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("err: %s url %s", "GET", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("error for %s url %s, %s", "GET", url, resp.Status)
	}
	return resp
}

//nolint:unused
func invokePost(client *http.Client, log *log.Logger, url string, json string) *http.Response {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(json)))
	req.Header.Set("Authorization", bearer)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("POST %s with body %s", url, json)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("err: %s url %s", "POST", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("error for %s url %s, %s with body %s", "POST", url, resp.Status, respBody)
	}
	return resp
}

//nolint:unused
func invokeDelete(client *http.Client, log *log.Logger, url string) *http.Response {
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("Authorization", bearer)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("err: %s url %s", "DELETE", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Fatalf("error for %s url %s, %s with body %s", "DELETE", url, resp.Status, respBody)
	}
	return resp
}

//nolint:unused
func getOrganizationById(client *http.Client, log *log.Logger) {
	var apitype OrganizationResponse
	var url = ApiUrl + "/organizations/" + organizationId

	resp := invokeGet(client, log, url)
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(&apitype)
	if err != nil {
		log.Fatalf("Can't deser organizationResponse %s", err)
	}
	log.Printf("%#v", apitype)
}

//nolint:unused
func getProjectsList(client *http.Client, log *log.Logger) {
	var apitype []ProjectResponse
	var url = ApiUrl + "/organizations/" + organizationId + "/projects"

	resp := invokeGet(client, log, url)
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(&apitype)
	if err != nil {
		log.Fatalf("Can't deser ProjectResponse array %s", err)
	}
	log.Printf("%#v", apitype)
}

//nolint:unused
func createAndDeleteProject(client *http.Client, log *log.Logger) {
	// create
	var apitype ProjectResponse
	var url = ApiUrl + "/organizations/" + organizationId + "/projects"
	var body = `{"name":"rbettstestproject"}`

	resp := invokePost(client, log, url, body)
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(&apitype)
	if err != nil {
		log.Fatalf("Can't deser ProjectResponse %s", err)
	}
	log.Printf("%#v", apitype)

	// delete
	url2 := ApiUrl + "/projects/" + apitype.Id
	resp2 := invokeDelete(client, log, url2)
	defer resp2.Body.Close()
	log.Printf("Delete of newly created project response: %s", resp2.Status)
}
